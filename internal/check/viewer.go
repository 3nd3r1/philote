package check

import (
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// viewer resolves a frontend's viewer line. The app reads the viewer's
// record by id on every page that shows it, so its entity needs a resource
// with get in the backend.
func (c *checker) viewer(part, backend *ir.Part, s *syntax.Viewer) {
	ent := c.lookupEntity(s.Entity)
	if ent == nil {
		return
	}
	for _, r := range backend.Resources {
		if r.Entity == ent.Name && slices.Contains(r.Ops, ir.Get) {
			part.Viewer = &ir.Viewer{Entity: ent.Name, Resource: r.Name, Pos: s.Word.Pos}
			return
		}
	}
	c.d.Errorf(s.Entity.Pos, "the viewer needs a resource in %s that gets a %s by id", backend.Name, ent.Name)
	c.d.Hint("add `resource <name>: %s` to %s", ent.Name, backend.Name)
}

// viewerSection checks C(viewer): C takes one record of the viewer's entity.
func (c *checker) viewerSection(part *ir.Part, comp *ir.Component, sec *ir.Section, s syntax.Section) *ir.Section {
	v := part.Viewer
	switch {
	case v == nil:
		c.d.Errorf(s.Arg.Pos, "%s has no viewer", part.Name)
		c.d.Hint("add `viewer: %s` to %s", or(comp.Entity, "<Entity>"), part.Name)
	case comp.Param == "":
		c.d.Errorf(s.Arg.Pos, "%s takes no record", comp.Name)
	case comp.Entity != v.Entity:
		c.d.Errorf(s.Arg.Pos, "%s shows %s records, but the viewer is a %s", comp.Name, comp.Entity, v.Entity)
	case s.List || comp.List:
		c.d.Errorf(s.Component.Pos, "the viewer is one record; show it with a component that takes one %s", v.Entity)
	default:
		sec.Viewer = true
		return sec
	}
	return nil
}

// picks checks `picks: viewer then "<route>"`: the component lists the
// records a visitor can be, and tapping one makes it the viewer.
func (c *checker) picks(part *ir.Part, comp *ir.Component, s *syntax.Component) {
	if s.Picks == nil {
		return
	}
	v, then := part.Viewer, s.PicksThen.Value
	switch {
	case s.Picks.Name != "viewer":
		c.d.Errorf(s.Picks.Pos, "a component can pick only the viewer, like picks: viewer then \"/\"")
	case v == nil:
		c.d.Errorf(s.Picks.Pos, "%s has no viewer to pick", part.Name)
		c.d.Hint("add `viewer: <Entity>` to %s", part.Name)
	case !comp.List || comp.Entity != v.Entity:
		c.d.Errorf(comp.Pos, "%s picks a %s, so it takes the list to pick from", comp.Name, v.Entity)
		c.d.Hint("give it a parameter like (%ss: list of %s)", strings.ToLower(v.Entity), v.Entity)
	case findPage(part, then) == nil:
		c.d.Errorf(s.PicksThen.Pos, "%s has no page at %s", part.Name, then)
		c.d.HintNames(then, navRoutes(part), "pages")
	case strings.Contains(then, "["):
		c.d.Errorf(s.PicksThen.Pos, "picking can't go to %s: its route needs an id", then)
	default:
		comp.Picks = then
	}
}

// picker finds the one page that picks the viewer, where a page that needs
// the viewer sends a visitor who hasn't picked yet.
func (c *checker) picker(part *ir.Part) {
	v := part.Viewer
	if v == nil {
		return
	}
	var pickers, needers []string
	for _, pg := range part.Pages {
		if slices.ContainsFunc(pg.Sections, func(s *ir.Section) bool { return findComponent(part, s.Component).Picks != "" }) {
			pickers = append(pickers, pg.Route)
		}
		if pg.Viewer {
			needers = append(needers, pg.Route)
		}
	}
	// The action behind a call as the viewer sends a visitor without one to
	// the picker too, even from a component no page shows yet.
	for _, comp := range part.Components {
		if comp.Calls != nil && len(comp.Calls.Viewer) > 0 {
			needers = append(needers, comp.Name)
		}
	}
	switch {
	case len(pickers) == 0 && len(needers) > 0:
		c.d.Errorf(v.Pos, "%s needs the viewer, but no page picks it", needers[0])
		c.d.Hint("add `picks: viewer then \"%s\"` to a component that takes a list of %s, and show it on a page", needers[0], v.Entity)
	case len(pickers) > 1:
		c.d.Errorf(v.Pos, "%s and %s both pick the viewer; keep one picker page", pickers[0], pickers[1])
	case len(pickers) == 1 && slices.Contains(needers, pickers[0]):
		c.d.Errorf(v.Pos, "%s picks the viewer, so it can't need it too", pickers[0])
	case len(pickers) == 1:
		v.Picker = pickers[0]
	}
}
