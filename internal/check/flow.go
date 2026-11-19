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

// value resolves an input, parameter or result type: a scalar or an entity
// record. Only one record can be optional, since a list can just be empty.
func (c *checker) value(f *syntax.Field) *ir.Value {
	v := &ir.Value{Name: f.Name.Name, List: f.List, Optional: len(f.Mods) > 0}
	if v.Optional && (f.List || c.entities[f.Type.Name] == nil) {
		c.d.Errorf(f.Mods[0].Pos, "only one record can be optional, like `match: Match optional`")
		return nil
	}
	if c.entities[f.Type.Name] != nil {
		if len(f.Limits) > 0 {
			c.d.Errorf(f.Limits[0].Name.Pos, "(min/max) only applies to string and int values")
		}
		v.Entity = f.Type.Name
		return v
	}
	if f.List {
		c.d.Errorf(f.Type.Pos, "only records come in a list, like list of Horse")
		return nil
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
		return a.Entity == b.Entity && a.List == b.List && a.Optional == b.Optional
	}
	return a.Field.Kind == b.Field.Kind && a.Field.Enum == b.Field.Enum
}

func typeName(v *ir.Value) string {
	switch {
	case v.List:
		return "list of " + v.Entity
	case v.Optional:
		return v.Entity + " optional"
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
		v := c.value(f)
		switch {
		case v == nil:
		case v.Optional:
			c.d.Errorf(f.Mods[0].Pos, "only a step's result or an endpoint's output can be optional")
		default:
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
	st.Reads = c.tables(s.Reads, nil)
	st.Writes = c.tables(s.Writes, st.Reads)
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

// tables resolves a step's reads or writes. A table read and written goes
// under writes only, so the lists never overlap (D-022).
func (c *checker) tables(names []syntax.Ident, reads []string) []string {
	var out []string
	for _, n := range names {
		switch {
		case slices.Contains(out, n.Name):
			c.d.Errorf(n.Pos, "'%s' is listed twice", n.Name)
		case slices.Contains(reads, n.Name):
			c.d.Errorf(n.Pos, "%s is under both reads and writes", n.Name)
			c.d.Hint("writes covers reading too; drop it from reads")
		case c.lookupEntity(n) != nil:
			out = append(out, n.Name)
		}
	}
	return out
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
	ep := &ir.Endpoint{Name: e.Name.Name, Flow: &ir.Flow{}}
	params, ok := c.route(part, e, ep)
	if !ok || !c.inputs(e, ep, params) {
		return
	}
	if e.Output != nil && c.lookupEntity(*e.Output) != nil {
		ep.Output, ep.List = e.Output.Name, e.List
	}
	outputs := c.outputs(e)
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
	scope := c.flow(part, e, ep)
	c.named(ep, outputs, scope)
	part.Custom = append(part.Custom, ep)
}

// route checks an endpoint's method and path, and returns the path's
// params in order, like horse for "/horses/[horse]/candidates".
func (c *checker) route(part *ir.Part, e *syntax.Endpoint, ep *ir.Endpoint) ([]string, bool) {
	if e.Method == nil {
		c.d.Errorf(e.Name.Pos, "endpoint %s needs a route", e.Name.Name)
		c.d.Hint(`add route: POST "/%s"`, e.Name.Name)
		return nil, false
	}
	if !slices.Contains(methods, e.Method.Name) {
		c.d.Errorf(e.Method.Pos, "unknown method '%s'", e.Method.Name)
		c.d.HintNames(e.Method.Name, methods, "methods")
		return nil, false
	}
	path := e.Path.Value
	segs := strings.Split(strings.TrimPrefix(path, "/"), "/")
	var params []string
	for _, seg := range segs {
		if name, ok := routeParam(seg); ok {
			params = append(params, name)
		} else if !isURLName(strings.ReplaceAll(seg, "-", "_")) {
			c.d.Errorf(e.Path.Pos, "a route is lowercase segments and [params], like \"/swipe\" or \"/horses/[horse]/candidates\"")
			return nil, false
		}
	}
	switch {
	case !strings.HasPrefix(path, "/"):
		c.d.Errorf(e.Path.Pos, "a route starts with /, like \"/swipe\"")
	case len(params) > 0 && e.Method.Name != "GET":
		c.d.Errorf(e.Path.Pos, "only a GET route takes [params]; send %s in the input", params[0])
	case c.resourceRoute(part, segs) != "":
		r := c.resourceRoute(part, segs)
		c.d.Errorf(e.Path.Pos, "%s is under the %s resource's routes", path, r)
		c.d.Hint("a GET route under a resource looks like \"/%s/[<input>]/<name>\"", r)
	default:
		ep.Method, ep.Route = e.Method.Name, path
		return params, true
	}
	return nil, false
}

// routeParam reads a segment like [horse].
func routeParam(seg string) (string, bool) {
	name, ok := strings.CutPrefix(seg, "[")
	name, closed := strings.CutSuffix(name, "]")
	return name, ok && closed && isURLName(name)
}

// resourceRoute names the resource whose routes would catch a custom route:
// any route under its path, except /<resource>/[<param>]/<name> when name
// isn't one of its relation reads.
func (c *checker) resourceRoute(part *ir.Part, segs []string) string {
	for _, r := range part.Resources {
		if segs[0] != r.Name {
			continue
		}
		if len(segs) != 3 || !strings.HasPrefix(segs[1], "[") {
			return r.Name
		}
		rel := findRelation(c.entities[r.Entity], segs[2])
		if rel != nil && rel.Many && slices.Contains(r.Ops, ir.Get) {
			return r.Name
		}
	}
	return ""
}

// inputs resolves an endpoint's input: records named in a GET route are
// sent there as ids, and a GET endpoint takes nothing else.
func (c *checker) inputs(e *syntax.Endpoint, ep *ir.Endpoint, params []string) bool {
	values := c.values(e.Input)
	for _, f := range e.Input {
		if f.List {
			c.d.Errorf(f.Name.Pos, "an input is one value; a list can't be sent")
			return false
		}
	}
	for _, name := range params {
		i := slices.IndexFunc(values, func(v *ir.Value) bool { return v.Name == name })
		switch {
		case i < 0:
			c.d.Errorf(e.Path.Pos, "[%s] isn't an input", name)
			c.d.Hint("add input: { %s: <Entity> }", name)
			return false
		case values[i].Entity == "":
			c.d.Errorf(e.Path.Pos, "[%s] isn't a record; a route takes only records, sent as their ids", name)
			return false
		}
		ep.Flow.Params = append(ep.Flow.Params, values[i])
	}
	if ep.Method != "GET" {
		ep.Flow.Input, ep.Errors = values, []string{"invalid_input"}
		return true
	}
	for _, f := range e.Input {
		if !slices.Contains(params, f.Name.Name) {
			c.d.Errorf(f.Name.Pos, "a GET endpoint takes no body; put %s in the route, like [%s]", f.Name.Name, f.Name.Name)
			return false
		}
	}
	if len(params) > 0 {
		ep.Errors = []string{"not_found"}
	}
	return true
}

// flow resolves the steps and checks that every parameter is bound and the
// errors line up.
func (c *checker) flow(part *ir.Part, e *syntax.Endpoint, ep *ir.Endpoint) map[string]*ir.Value {
	if len(e.Steps) == 0 {
		c.d.Errorf(e.Name.Pos, "endpoint %s needs steps", e.Name.Name)
		return nil
	}
	scope := map[string]*ir.Value{}
	for _, v := range slices.Concat(ep.Flow.Params, ep.Flow.Input) {
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
		if r := st.Result; r != nil && ep.Output != "" && r.Entity == ep.Output && r.List == ep.List && !r.Optional {
			if producer != "" {
				c.d.Errorf(ref.Pos, "%s and %s both return %s; exactly one step must", producer, st.Name, outputName(ep))
			}
			producer = st.Name
		}
		ep.Flow.Steps = append(ep.Flow.Steps, st)
	}
	if ep.Output != "" && producer == "" {
		c.d.Errorf(e.Output.Pos, "no step returns %s", outputName(ep))
	}
	for _, er := range e.Errors {
		if !raised[er.Name] && !slices.Contains(builtinErrors, er.Name) {
			c.d.Errorf(er.Pos, "no step can fail with %s", er.Name)
		}
	}
	return scope
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

// outputName is what an endpoint returns, like "a Horse" or "a list of Horse".
func outputName(ep *ir.Endpoint) string {
	if ep.List {
		return "a list of " + ep.Output
	}
	return "a " + ep.Output
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
