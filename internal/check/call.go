package check

import (
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// endpointLoad checks `load: <name> from <endpoint>(<input>: viewer, ...)`:
// a page loads the list a GET endpoint returns, with the viewer as its route
// inputs.
func (c *checker) endpointLoad(pg *ir.Page, part, backend *ir.Part, s *syntax.Page) bool {
	at := *s.Load.From
	ep := c.customEndpoint(backend, at)
	switch {
	case ep == nil:
		return false
	case ep.Method != "GET":
		c.d.Errorf(at.Pos, "%s is a %s endpoint; a page loads only from GET endpoints", ep.Name, ep.Method)
		c.d.Hint("a component can call it with `calls: %s(...)`", ep.Name)
		return false
	case ep.Output == "" || !ep.List:
		c.d.Errorf(at.Pos, "a page loads a list from an endpoint, but %s doesn't return one", ep.Name)
		return false
	case strings.Contains(pg.Route, "[id]"):
		c.d.Errorf(s.Route.Pos, "this page loads from %s, so its route can't have [id]", ep.Name)
		return false
	}
	args, ok := c.args(at, ep, ep.Flow.Params, s.Load.Args)
	if !ok {
		return false
	}
	for _, in := range ep.Flow.Params {
		a := args[in.Name]
		if a.Value == nil || a.Value.Name != "viewer" {
			c.d.Errorf(a.Name.Pos, "a page fills %s's inputs with the viewer, like %s: viewer", ep.Name, in.Name)
			return false
		}
		if !c.viewerArg(part, in, a) {
			return false
		}
		pg.Viewer = true
	}
	pg.Load, pg.Endpoint, pg.Entity = s.Load.Name.Name, ep.Name, ep.Output
	return true
}

// calls checks `calls: <endpoint>(<input>: viewer | <param>, <input>, ...)`:
// the template fills inputs bound to the viewer, and the component's body
// passes a record of its parameter or an open value for the rest.
func (c *checker) calls(part, backend *ir.Part, comp *ir.Component, s *syntax.Call) bool {
	ep := c.customEndpoint(backend, s.Endpoint)
	if ep == nil {
		return false
	}
	if ep.Method == "GET" {
		c.d.Errorf(s.Endpoint.Pos, "%s is a GET endpoint; a page loads it, and a component calls endpoints that change something", ep.Name)
		return false
	}
	args, ok := c.args(s.Endpoint, ep, ep.Flow.Input, s.Args)
	if !ok {
		return false
	}
	call := &ir.Call{Endpoint: ep.Name}
	for _, in := range ep.Flow.Input {
		a := args[in.Name]
		switch {
		case a.Value == nil && in.Entity != "":
			c.d.Errorf(a.Name.Pos, "%s is a %s, which the body can't make up; fill it with viewer or %s", in.Name, in.Entity, or(comp.Param, "a parameter"))
			return false
		case a.Value == nil:
			call.Passed = append(call.Passed, in)
		case a.Value.Name == "viewer":
			if !c.viewerArg(part, in, a) {
				return false
			}
			call.Viewer = append(call.Viewer, in.Name)
		case a.Value.Name != comp.Param || comp.Param == "":
			c.d.Errorf(a.Value.Pos, "'%s' isn't the viewer or %s's parameter", a.Value.Name, comp.Name)
			return false
		case comp.Entity != in.Entity:
			c.d.Errorf(a.Value.Pos, "%s holds %s records, but %s is a %s", comp.Param, comp.Entity, in.Name, typeName(in))
			return false
		default:
			call.Passed = append(call.Passed, in)
		}
	}
	comp.Calls = call
	return true
}

// customEndpoint finds an endpoint declared with steps in the backend.
func (c *checker) customEndpoint(backend *ir.Part, at syntax.Ident) *ir.Endpoint {
	i := slices.IndexFunc(backend.Custom, func(e *ir.Endpoint) bool { return e.Name == at.Name })
	if i >= 0 {
		return backend.Custom[i]
	}
	var names []string
	for _, e := range backend.Custom {
		names = append(names, e.Name)
	}
	c.d.Errorf(at.Pos, "%s has no endpoint named %s", backend.Name, at.Name)
	c.d.HintNames(at.Name, names, "endpoints")
	return nil
}

// args matches arguments to an endpoint's inputs: each input exactly once.
func (c *checker) args(at syntax.Ident, ep *ir.Endpoint, inputs []*ir.Value, args []syntax.Arg) (map[string]syntax.Arg, bool) {
	out := map[string]syntax.Arg{}
	var names []string
	for _, in := range inputs {
		names = append(names, in.Name)
	}
	for _, a := range args {
		switch {
		case !slices.Contains(names, a.Name.Name):
			c.d.Errorf(a.Name.Pos, "%s has no input named %s", ep.Name, a.Name.Name)
			c.d.HintNames(a.Name.Name, names, "inputs")
			return nil, false
		case out[a.Name.Name].Name.Name != "":
			c.d.Errorf(a.Name.Pos, "'%s' is listed twice", a.Name.Name)
			return nil, false
		}
		out[a.Name.Name] = a
	}
	for _, n := range names {
		if _, ok := out[n]; !ok {
			c.d.Errorf(at.Pos, "%s needs %s", ep.Name, n)
			c.d.Hint("list every input, like %s(%s)", ep.Name, strings.Join(names, ", "))
			return nil, false
		}
	}
	return out, true
}

// viewerArg checks that the viewer can fill an input.
func (c *checker) viewerArg(part *ir.Part, in *ir.Value, a syntax.Arg) bool {
	v := part.Viewer
	switch {
	case v == nil:
		c.d.Errorf(a.Value.Pos, "%s has no viewer", part.Name)
		c.d.Hint("add `viewer: %s` to %s", or(in.Entity, "<Entity>"), part.Name)
	case in.Entity != v.Entity || in.List:
		c.d.Errorf(a.Value.Pos, "the viewer is a %s, but %s is a %s", v.Entity, in.Name, typeName(in))
	default:
		return true
	}
	return false
}
