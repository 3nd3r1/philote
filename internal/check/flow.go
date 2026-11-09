package check

import (
	"slices"
	"sort"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

const maxIntent = 200

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// Errors the templates raise themselves; steps and endpoints can't declare them.
var builtinErrors = []string{"not_found", "invalid_input"}

// value resolves an input, parameter or result type: a scalar or an entity record.
func (c *checker) value(f *syntax.Field) *ir.Value {
	v := &ir.Value{Name: f.Name.Name}
	if c.entities[f.Type.Name] != nil {
		if len(f.Limits) > 0 {
			c.d.Errorf(f.Limits[0].Name.Pos, "(min/max) only applies to string and int values")
		}
		v.Entity = f.Type.Name
		return v
	}
	fd := &ir.Field{Name: f.Name.Name}
	kind, primitive := primitives[f.Type.Name]
	switch {
	case primitive:
		fd.Kind = kind
	case c.app.Enum(f.Type.Name) != nil:
		fd.Kind, fd.Enum = ir.EnumValue, f.Type.Name
	default:
		c.d.Errorf(f.Type.Pos, "unknown type '%s'", f.Type.Name)
		names := append(slices.Clone(primitiveNames), enumNames(c.app)...)
		c.d.HintNames(f.Type.Name, append(names, c.entityNames()...), "types")
		return nil
	}
	c.limits(fd, f)
	v.Field = fd
	return v
}

func sameType(a, b *ir.Value) bool {
	if a.Entity != "" || b.Entity != "" {
		return a.Entity == b.Entity
	}
	return a.Field.Kind == b.Field.Kind && a.Field.Enum == b.Field.Enum
}

func typeName(v *ir.Value) string {
	switch {
	case v.Entity != "":
		return v.Entity
	case v.Field.Kind == ir.EnumValue:
		return v.Field.Enum
	}
	return primitiveNames[v.Field.Kind]
}

// values resolves a list of fields, reporting duplicate names.
func (c *checker) values(fields []*syntax.Field) []*ir.Value {
	var out []*ir.Value
	seen := map[string]bool{}
	for _, f := range fields {
		if seen[f.Name.Name] {
			c.d.Errorf(f.Name.Pos, "'%s' is listed twice", f.Name.Name)
			continue
		}
		seen[f.Name.Name] = true
		if v := c.value(f); v != nil {
			out = append(out, v)
		}
	}
	return out
}

func (c *checker) step(part *ir.Part, s *syntax.Step) {
	if !isURLName(s.Name.Name) {
		c.d.Errorf(s.Name.Pos, "step names must be lowercase, like `record_swipe`")
		return
	}
	if slices.ContainsFunc(part.Steps, func(x *ir.Step) bool { return x.Name == s.Name.Name }) {
		c.d.Errorf(s.Name.Pos, "%s already has a step named %s", part.Name, s.Name.Name)
		return
	}
	st := &ir.Step{Name: s.Name.Name, Pos: s.Name.Pos, Params: c.values(s.Params)}
	if s.Result != nil {
		if slices.ContainsFunc(s.Params, func(f *syntax.Field) bool { return f.Name.Name == s.Result.Name.Name }) {
			c.d.Errorf(s.Result.Name.Pos, "the result can't have the same name as a parameter")
		}
		st.Result = c.value(s.Result)
	}
	for _, f := range s.Fails {
		switch {
		case f.Name == "invalid_input":
			c.d.Errorf(f.Pos, "invalid_input is raised by input checking, not by steps")
		case slices.Contains(st.Fails, f.Name):
			c.d.Errorf(f.Pos, "'%s' is listed twice", f.Name)
		default:
			st.Fails = append(st.Fails, f.Name)
		}
	}
	switch {
	case s.Intent == nil:
		c.d.Errorf(s.Name.Pos, "step %s needs an intent", st.Name)
		c.d.Hint(`add { intent: "what the step does, in one sentence" }`)
	case len(s.Intent.Value) > maxIntent:
		c.d.Errorf(s.Intent.Pos, "the intent is %d characters; the cap is %d", len(s.Intent.Value), maxIntent)
	default:
		st.Intent = s.Intent.Value
	}
	part.Steps = append(part.Steps, st)
}

func (c *checker) endpoint(part *ir.Part, e *syntax.Endpoint) {
	if !isURLName(e.Name.Name) {
		c.d.Errorf(e.Name.Pos, "endpoint names must be lowercase, like `swipe`")
		return
	}
	if slices.ContainsFunc(part.Custom, func(x *ir.Endpoint) bool { return x.Name == e.Name.Name }) {
		c.d.Errorf(e.Name.Pos, "%s already has an endpoint named %s", part.Name, e.Name.Name)
		return
	}
	ep := &ir.Endpoint{Name: e.Name.Name, Errors: []string{"invalid_input"}, Flow: &ir.Flow{}}
	if !c.route(part, e, ep) {
		return
	}
	ep.Flow.Input = c.values(e.Input)
	if ep.Method == "GET" && len(e.Input) > 0 {
		c.d.Errorf(e.Input[0].Name.Pos, "GET endpoints can't take an input body")
	}
	if e.Output != nil && c.lookupEntity(*e.Output) != nil {
		ep.Output = e.Output.Name
	}
	for _, er := range e.Errors {
		switch {
		case slices.Contains(builtinErrors, er.Name):
			c.d.Errorf(er.Pos, "%s is built in; don't list it", er.Name)
		case slices.Contains(ep.Errors, er.Name):
			c.d.Errorf(er.Pos, "'%s' is listed twice", er.Name)
		default:
			ep.Errors = append(ep.Errors, er.Name)
		}
	}
	c.flow(part, e, ep)
	part.Custom = append(part.Custom, ep)
}

func (c *checker) route(part *ir.Part, e *syntax.Endpoint, ep *ir.Endpoint) bool {
	if e.Method == nil {
		c.d.Errorf(e.Name.Pos, "endpoint %s needs a route", e.Name.Name)
		c.d.Hint(`add route: POST "/%s"`, e.Name.Name)
		return false
	}
	if !slices.Contains(methods, e.Method.Name) {
		c.d.Errorf(e.Method.Pos, "unknown method '%s'", e.Method.Name)
		c.d.HintNames(e.Method.Name, methods, "methods")
		return false
	}
	path := e.Path.Value
	segs := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if !strings.HasPrefix(path, "/") || slices.ContainsFunc(segs, func(s string) bool { return !isURLName(strings.ReplaceAll(s, "-", "_")) }) {
		c.d.Errorf(e.Path.Pos, "a route is lowercase segments like \"/swipe\"; params aren't supported here yet")
		return false
	}
	for _, r := range part.Resources {
		if segs[0] == r.Name {
			c.d.Errorf(e.Path.Pos, "%s is under the %s resource's routes", path, r.Name)
			return false
		}
	}
	ep.Method, ep.Route = e.Method.Name, path
	return true
}

// flow resolves the steps and checks that every parameter is bound and the
// errors line up.
func (c *checker) flow(part *ir.Part, e *syntax.Endpoint, ep *ir.Endpoint) {
	if len(e.Steps) == 0 {
		c.d.Errorf(e.Name.Pos, "endpoint %s needs steps", e.Name.Name)
		return
	}
	scope := map[string]*ir.Value{}
	for _, v := range ep.Flow.Input {
		scope[v.Name] = v
	}
	var producer string
	raised := map[string]bool{}
	for _, ref := range e.Steps {
		i := slices.IndexFunc(part.Steps, func(s *ir.Step) bool { return s.Name == ref.Name })
		if i < 0 {
			c.d.Errorf(ref.Pos, "unknown step '%s'", ref.Name)
			c.d.HintNames(ref.Name, stepNames(part), "steps")
			continue
		}
		st := part.Steps[i]
		c.bind(ref, st, scope)
		c.fails(ref, st, ep, raised)
		if st.Result != nil && ep.Output != "" && st.Result.Entity == ep.Output {
			if producer != "" {
				c.d.Errorf(ref.Pos, "%s and %s both return a %s; exactly one step must", producer, st.Name, ep.Output)
			}
			producer = st.Name
		}
		ep.Flow.Steps = append(ep.Flow.Steps, st)
	}
	if ep.Output != "" && producer == "" {
		c.d.Errorf(e.Output.Pos, "no step returns a %s", ep.Output)
	}
	for _, er := range e.Errors {
		if !raised[er.Name] && !slices.Contains(builtinErrors, er.Name) {
			c.d.Errorf(er.Pos, "no step can fail with %s", er.Name)
		}
	}
}

// fails checks that every error a step can raise is declared on the endpoint.
func (c *checker) fails(ref syntax.Ident, st *ir.Step, ep *ir.Endpoint, raised map[string]bool) {
	for _, f := range st.Fails {
		raised[f] = true
		if !slices.Contains(ep.Errors, f) && f != "not_found" {
			c.d.Errorf(ref.Pos, "step %s can fail with %s, which endpoint %s doesn't list in errors", st.Name, f, ep.Name)
		}
	}
}

// bind checks a step's parameters against what's in scope, then adds its result.
func (c *checker) bind(ref syntax.Ident, st *ir.Step, scope map[string]*ir.Value) {
	for _, p := range st.Params {
		v := scope[p.Name]
		switch {
		case v == nil:
			c.d.Errorf(ref.Pos, "step %s needs '%s', but no input or earlier step provides it", st.Name, p.Name)
			c.d.Hint("available here: %s", strings.Join(sortedKeys(scope), ", "))
		case !sameType(v, p):
			c.d.Errorf(ref.Pos, "'%s' is a %s here, but step %s takes a %s", p.Name, typeName(v), st.Name, typeName(p))
		}
	}
	if r := st.Result; r != nil {
		if scope[r.Name] != nil {
			c.d.Errorf(ref.Pos, "step %s returns '%s', which is already defined here", st.Name, r.Name)
		}
		scope[r.Name] = r
	}
}

func stepNames(part *ir.Part) []string {
	var out []string
	for _, s := range part.Steps {
		out = append(out, s.Name)
	}
	return out
}

func sortedKeys(m map[string]*ir.Value) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
