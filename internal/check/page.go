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
	case route == "/":
		c.d.Errorf(s.Route.Pos, "/ is the base project's home page; pick another route")
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
	if !c.pageData(pg, backend, s) || !c.pageLayout(pg, part, s) || !c.pageCopy(pg, s) {
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

func (c *checker) pageData(pg *ir.Page, backend *ir.Part, s *syntax.Page) bool {
	if s.Data == nil {
		c.d.Errorf(s.Word.Pos, "page %s needs data", pg.Route)
		c.d.Hint("add `data: <resource>`, a resource of %s", backend.Name)
		return false
	}
	res := findResource(backend, s.Data.Name)
	if res == nil {
		c.d.Errorf(s.Data.Pos, "%s has no resource named %s", backend.Name, s.Data.Name)
		c.d.HintNames(s.Data.Name, resourceNames(backend), "resources")
		return false
	}
	pg.Resource, pg.Entity = res.Name, res.Entity
	if s.By != nil && s.By.Name != "id" {
		c.d.Errorf(s.By.Pos, "pages load by id only, for now")
		return false
	}
	pg.ByID = s.By != nil
	op, hasID := ir.List, strings.Contains(pg.Route, "[id]")
	switch {
	case pg.ByID && !hasID:
		c.d.Errorf(s.Route.Pos, "a page that loads by id needs [id] in its route")
		return false
	case !pg.ByID && hasID:
		c.d.Errorf(s.Data.Pos, "this page lists %s, but its route has [id]", res.Name)
		c.d.Hint("load one record with `data: %s by id`", res.Name)
		return false
	case pg.ByID:
		op = ir.Get
	}
	if !slices.Contains(res.Ops, op) {
		c.d.Errorf(s.Data.Pos, "resource %s doesn't serve %s", res.Name, op)
		return false
	}
	return true
}

func (c *checker) pageLayout(pg *ir.Page, part *ir.Part, s *syntax.Page) bool {
	if s.Layout == nil {
		c.d.Errorf(s.Word.Pos, "page %s needs a layout", pg.Route)
		c.d.Hint("add `layout: <Component>`")
		return false
	}
	comp := findComponent(part, s.Layout.Name)
	switch {
	case comp == nil:
		var names []string
		for _, x := range part.Components {
			names = append(names, x.Name)
		}
		c.d.Errorf(s.Layout.Pos, "unknown component '%s'", s.Layout.Name)
		c.d.HintNames(s.Layout.Name, names, "components")
	case !pg.ByID && !s.List:
		c.d.Errorf(s.Layout.Pos, "this page lists records; use `layout: list of %s`", comp.Name)
	case pg.ByID && s.List:
		c.d.Errorf(s.Layout.Pos, "this page loads one record; use `layout: %s`", comp.Name)
	case comp.Entity != pg.Entity:
		c.d.Errorf(s.Layout.Pos, "%s shows a %s, but this page loads %s records", comp.Name, comp.Entity, pg.Entity)
	default:
		pg.Component = comp.Name
		return true
	}
	return false
}

// pageCopy checks D-006: a list page says what shows when it's empty, a
// page by id says what shows when the record doesn't exist.
func (c *checker) pageCopy(pg *ir.Page, s *syntax.Page) bool {
	ok := true
	for _, str := range []*syntax.Str{s.Empty, s.NotFound} {
		if str != nil && len(str.Value) > maxCopy {
			c.d.Errorf(str.Pos, "UI text is capped at %d characters", maxCopy)
			ok = false
		}
	}
	switch {
	case !pg.ByID && s.Empty == nil:
		c.d.Errorf(s.Word.Pos, "a list page must say what shows when it's empty")
		c.d.Hint(`add empty: "No %s yet."`, pg.Resource)
	case pg.ByID && s.NotFound == nil:
		c.d.Errorf(s.Word.Pos, "a page by id must say what shows when the record doesn't exist")
		c.d.Hint(`add not_found: "That %s doesn't exist."`, strings.ToLower(pg.Entity))
	case pg.ByID && s.Empty != nil:
		c.d.Errorf(s.Empty.Pos, "only a list page can be empty")
	case !pg.ByID && s.NotFound != nil:
		c.d.Errorf(s.NotFound.Pos, "only a page by id has not_found")
	default:
		if pg.ByID {
			pg.NotFound = s.NotFound.Value
		} else {
			pg.Empty = s.Empty.Value
		}
		return ok
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
