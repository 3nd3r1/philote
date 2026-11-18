package check

import (
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// page checks one page and adds it to part. For a form page it also
// returns the entity a save returns, which its `then` resolves against
// once every page is known.
func (c *checker) page(part, backend *ir.Part, s *syntax.Page) (*ir.Page, string) {
	route := s.Route.Value
	switch {
	case !isRoute(route):
		c.d.Errorf(s.Route.Pos, `a route is lowercase segments like "/horses" or "/horses/[id]"`)
		return nil, ""
	case findPage(part, route) != nil:
		c.d.Errorf(s.Route.Pos, "there is already a page at %s", route)
		return nil, ""
	}
	pg := &ir.Page{Route: route}
	if s.Form != nil {
		saved, ok := c.formPage(pg, backend, s)
		if !ok {
			return nil, ""
		}
		part.Pages = append(part.Pages, pg)
		return pg, saved
	}
	if s.Then != nil {
		c.d.Errorf(s.Then.Pos, "only a form page has then")
		return nil, ""
	}
	if !c.pageLoad(pg, part, backend, s) || !c.pageSections(pg, part, s) || !c.pageCopy(pg, s) {
		return nil, ""
	}
	part.Pages = append(part.Pages, pg)
	return pg, ""
}

// isRoute accepts /a/b-c/[id]: lowercase segments, and [id] as the only param.
func isRoute(route string) bool {
	if !strings.HasPrefix(route, "/") {
		return false
	}
	if route == "/" {
		return true
	}
	for _, seg := range strings.Split(route[1:], "/") {
		if seg != "[id]" && !isURLName(strings.ReplaceAll(seg, "-", "_")) {
			return false
		}
	}
	return true
}

// pageLoad resolves what a page fetches: a resource's list, one record by
// the route's [id], or what a GET endpoint returns. A page without load
// fetches nothing.
func (c *checker) pageLoad(pg *ir.Page, part, backend *ir.Part, s *syntax.Page) bool {
	hasID := strings.Contains(pg.Route, "[id]")
	if s.Load == nil {
		if hasID {
			c.d.Errorf(s.Route.Pos, "this route has [id], but the page loads nothing")
			c.d.Hint("add `load: <name> from <resource> by id`")
			return false
		}
		return true
	}
	if s.Load.Name.Name == "viewer" {
		c.d.Errorf(s.Load.Name.Pos, "viewer is the app's viewer; name what the page loads something else")
		c.d.Hint("show the viewer with `shows: <Component>(viewer)`")
		return false
	}
	if s.Load.Call {
		return c.endpointLoad(pg, part, backend, s)
	}
	return c.resourceLoad(pg, backend, s, hasID)
}

// resourceLoad resolves a load from a resource: its list, or one record by
// the route's [id].
func (c *checker) resourceLoad(pg *ir.Page, backend *ir.Part, s *syntax.Page, hasID bool) bool {
	from := s.Load.Name
	if s.Load.From != nil {
		from = *s.Load.From
	}
	res := findResource(backend, from.Name)
	if res == nil {
		c.d.Errorf(from.Pos, "%s has no resource named %s", backend.Name, from.Name)
		c.d.HintNames(from.Name, resourceNames(backend), "resources")
		return false
	}
	pg.Load, pg.Resource, pg.Entity = s.Load.Name.Name, res.Name, res.Entity
	if s.Load.By != nil && s.Load.By.Name != "id" {
		c.d.Errorf(s.Load.By.Pos, "pages load by id only, for now")
		return false
	}
	pg.ByID = s.Load.By != nil
	op := ir.List
	switch {
	case pg.ByID && !hasID:
		c.d.Errorf(s.Route.Pos, "a page that loads by id needs [id] in its route")
		return false
	case !pg.ByID && hasID:
		c.d.Errorf(from.Pos, "this page lists %s, but its route has [id]", res.Name)
		c.d.Hint("load one record with `load: %s by id`", s.Load.Name.Name)
		return false
	case pg.ByID:
		op = ir.Get
	}
	if !slices.Contains(res.Ops, op) {
		c.d.Errorf(from.Pos, "resource %s doesn't serve %s", res.Name, op)
		return false
	}
	return true
}

// pageSections resolves what a page shows, in order. A section's argument
// is what the page loads, and its component must take that shape.
func (c *checker) pageSections(pg *ir.Page, part *ir.Part, s *syntax.Page) bool {
	if len(s.Shows) == 0 {
		c.d.Errorf(s.Word.Pos, "page %s doesn't show anything", pg.Route)
		c.d.Hint("add `shows: <Component>`")
		return false
	}
	ok := true
	for _, sec := range s.Shows {
		x := c.section(pg, part, sec)
		if x == nil {
			ok = false
			continue
		}
		pg.Sections = append(pg.Sections, x)
		if calls := findComponent(part, x.Component).Calls; x.Viewer || calls != nil && len(calls.Viewer) > 0 {
			pg.Viewer = true
		}
	}
	if pg.Load != "" && !slices.ContainsFunc(pg.Sections, func(x *ir.Section) bool { return x.Arg && !x.Viewer }) {
		c.d.Errorf(s.Load.Name.Pos, "the page loads %s but shows nothing with it", pg.Load)
		c.d.Hint("pass it to a component, like `shows: <Component>(%s)`", pg.Load)
		return false
	}
	return ok
}

func (c *checker) section(pg *ir.Page, part *ir.Part, s syntax.Section) *ir.Section {
	comp := findComponent(part, s.Component.Name)
	if comp == nil {
		c.d.Errorf(s.Component.Pos, "unknown component '%s'", s.Component.Name)
		c.d.HintNames(s.Component.Name, componentNames(part), "components")
		return nil
	}
	sec := &ir.Section{Component: comp.Name, List: s.List, Arg: s.Arg != nil}
	switch {
	case s.Arg == nil && comp.Param != "":
		c.d.Errorf(s.Component.Pos, "%s takes %s; pass it what the page loads, like %s(%s)", comp.Name, comp.Param, comp.Name, or(pg.Load, "<load>"))
	case s.Arg == nil && s.List:
		c.d.Errorf(s.Component.Pos, "a list of %s needs a list to show, like list of %s(%s)", comp.Name, comp.Name, or(pg.Load, "<load>"))
	case s.Arg == nil:
		return sec
	case s.Arg.Name == "viewer":
		return c.viewerSection(part, comp, sec, s)
	case s.Arg.Name != pg.Load:
		c.d.Errorf(s.Arg.Pos, "the page loads nothing named %s", s.Arg.Name)
		if pg.Load != "" {
			c.d.Hint("it loads %s", pg.Load)
		}
	case comp.Param == "":
		c.d.Errorf(s.Arg.Pos, "%s takes no record", comp.Name)
	case comp.Entity != pg.Entity:
		c.d.Errorf(s.Arg.Pos, "%s shows %s records, but %s holds %s records", comp.Name, comp.Entity, pg.Load, pg.Entity)
	default:
		return c.sectionShape(pg, comp, sec, s)
	}
	return nil
}

// sectionShape checks one record against a list: `list of C(xs)` shows one
// C per record, and C(x) takes x as it is.
func (c *checker) sectionShape(pg *ir.Page, comp *ir.Component, sec *ir.Section, s syntax.Section) *ir.Section {
	switch {
	case s.List && pg.ByID:
		c.d.Errorf(s.Component.Pos, "%s is one record; show it with %s(%s)", pg.Load, comp.Name, pg.Load)
	case s.List && comp.List:
		c.d.Errorf(s.Component.Pos, "%s already takes a list; show it with %s(%s)", comp.Name, comp.Name, pg.Load)
	case !s.List && pg.ByID && comp.List:
		c.d.Errorf(s.Arg.Pos, "%s takes a list, but %s is one record", comp.Name, pg.Load)
	case !s.List && !pg.ByID && !comp.List:
		c.d.Errorf(s.Arg.Pos, "%s is a list and %s takes one record; use list of %s(%s)", pg.Load, comp.Name, comp.Name, pg.Load)
	default:
		return sec
	}
	return nil
}

func or(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func componentNames(part *ir.Part) []string {
	var names []string
	for _, x := range part.Components {
		names = append(names, x.Name)
	}
	return names
}

// pageCopy checks D-006: a page that loads a list says what shows when
// it's empty, a page by id says what shows when the record doesn't exist.
func (c *checker) pageCopy(pg *ir.Page, s *syntax.Page) bool {
	for _, str := range []*syntax.Str{s.Empty, s.NotFound} {
		if str != nil && len(str.Value) > maxCopy {
			c.d.Errorf(str.Pos, "UI text is capped at %d characters", maxCopy)
			return false
		}
	}
	list := pg.Load != "" && !pg.ByID
	switch {
	case list && s.Empty == nil:
		c.d.Errorf(s.Word.Pos, "a page that loads a list must say what shows when it's empty")
		c.d.Hint(`add empty: "No %s yet."`, pg.Load)
	case pg.ByID && s.NotFound == nil:
		c.d.Errorf(s.Word.Pos, "a page by id must say what shows when the record doesn't exist")
		c.d.Hint(`add not_found: "That %s doesn't exist."`, strings.ToLower(pg.Entity))
	case !list && s.Empty != nil:
		c.d.Errorf(s.Empty.Pos, "only a page that loads a list can be empty")
	case !pg.ByID && s.NotFound != nil:
		c.d.Errorf(s.NotFound.Pos, "only a page by id has not_found")
	default:
		if s.Empty != nil {
			pg.Empty = s.Empty.Value
		}
		if s.NotFound != nil {
			pg.NotFound = s.NotFound.Value
		}
		return true
	}
	return false
}

// link resolves links_to.
func (c *checker) link(part *ir.Part, comp *ir.Component, s *syntax.Str) {
	comp.LinksTo, comp.LinkBy, _ = c.recordRoute(part, comp, s)
}

// recordRoute resolves a route filled from a component's record:
// "/horses/[horse.id]" names the page at "/horses/[id]", and the component
// fills [id] with horse.id.
func (c *checker) recordRoute(part *ir.Part, comp *ir.Component, s *syntax.Str) (string, []string, bool) {
	ent := c.entities[comp.Entity]
	segs := strings.Split(s.Value, "/")
	var by []string
	for i, seg := range segs {
		name, ok := strings.CutPrefix(seg, "["+comp.Param+".")
		if !ok || !strings.HasSuffix(name, "]") {
			continue
		}
		field := strings.TrimSuffix(name, "]")
		if field != "id" && findField(ent, field) == nil {
			c.d.Errorf(s.Pos, "%s has no field '%s' to fill the link", comp.Entity, field)
			return "", nil, false
		}
		segs[i], by = "["+field+"]", append(by, field)
	}
	route := strings.Join(segs, "/")
	if findPage(part, route) == nil {
		c.d.Errorf(s.Pos, "there is no page at %s", route)
		c.d.HintNames(route, pageRoutes(part), "pages")
		return "", nil, false
	}
	return route, by, true
}
