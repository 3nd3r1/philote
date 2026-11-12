package check

import (
	"slices"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// actions resolves a component's actions once every page is known (D-020).
func (c *checker) actions(part, backend *ir.Part, comp *ir.Component, src []syntax.Action) {
	for _, s := range src {
		a := &ir.Action{}
		ok := c.actionKind(part, backend, comp, a, s)
		if ok && s.Then != nil {
			a.Then, a.ThenBy, ok = c.recordRoute(part, comp, s.Then)
		}
		if ok && s.Label != nil {
			if len(s.Label.Value) > maxCopy {
				c.d.Errorf(s.Label.Pos, "UI text is capped at %d characters", maxCopy)
				continue
			}
			a.Label = s.Label.Value
		}
		if ok {
			comp.Actions = append(comp.Actions, a)
		}
	}
}

func (c *checker) actionKind(part, backend *ir.Part, comp *ir.Component, a *ir.Action, s syntax.Action) bool {
	switch s.Verb.Name {
	case "open":
		if s.Then != nil {
			c.d.Errorf(s.Then.Pos, "open already goes to a page; then doesn't apply")
			return false
		}
		var ok bool
		a.Kind = ir.Open
		a.Route, a.RouteBy, ok = c.recordRoute(part, comp, s.Route)
		return ok
	case "delete":
		if slices.ContainsFunc(comp.Actions, func(x *ir.Action) bool { return x.Kind == ir.Remove }) {
			c.d.Errorf(s.Verb.Pos, "%s already has a delete", comp.Name)
			return false
		}
		res := c.serving(backend, comp.Entity, ir.Delete, "deletes", s.Verb)
		a.Kind, a.Endpoint = ir.Remove, res+".delete"
		return res != ""
	}
	c.d.Errorf(s.Verb.Pos, "unknown action '%s'", s.Verb.Name)
	c.d.HintNames(s.Verb.Name, []string{"open", "delete"}, "actions")
	return false
}
