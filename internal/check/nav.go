package check

import (
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// navPlaces are where a nav can sit.
var navPlaces = []string{"top", "bottom", "side"}

// maxNavLabel keeps an item short enough for a phone's tab bar.
const maxNavLabel = 24

// nav checks a frontend's nav line against its pages, which are all known
// by now.
func (c *checker) nav(part *ir.Part, s *syntax.Nav) {
	if !slices.Contains(navPlaces, s.Place.Name) {
		c.d.Errorf(s.Place.Pos, "a nav sits at the top, bottom or side, not %s", s.Place.Name)
		c.d.HintNames(s.Place.Name, navPlaces, "places")
		return
	}
	if len(s.Items) == 0 {
		c.d.Errorf(s.Word.Pos, "the nav has no pages")
		c.d.Hint(`list them after the place, like nav: %s, "/horses"`, s.Place.Name)
		return
	}
	nav := &ir.Nav{Place: s.Place.Name, Pos: s.Word.Pos}
	for _, it := range s.Items {
		item := c.navItem(part, nav, it)
		if item == nil {
			return
		}
		nav.Items = append(nav.Items, item)
	}
	part.Nav = nav
}

func (c *checker) navItem(part *ir.Part, nav *ir.Nav, it syntax.NavItem) *ir.NavItem {
	route := it.Route.Value
	switch {
	case findPage(part, route) == nil:
		c.d.Errorf(it.Route.Pos, "%s has no page at %s", part.Name, route)
		c.d.HintNames(route, navRoutes(part), "pages")
		return nil
	case strings.Contains(route, "["):
		c.d.Errorf(it.Route.Pos, "the nav can't link to %s: its route needs an id", route)
		return nil
	case slices.ContainsFunc(nav.Items, func(x *ir.NavItem) bool { return x.Route == route }):
		c.d.Errorf(it.Route.Pos, "%s is already in the nav", route)
		return nil
	}
	item := &ir.NavItem{Route: route}
	if it.Label != nil {
		switch {
		case it.Label.Value == "":
			c.d.Errorf(it.Label.Pos, "this text is empty")
			c.d.Hint("leave out `as` to use the page's title")
			return nil
		case len(it.Label.Value) > maxNavLabel:
			c.d.Errorf(it.Label.Pos, "a nav item's text is capped at %d characters", maxNavLabel)
			return nil
		}
		item.Label = it.Label.Value
	}
	return item
}

// navRoutes are the pages a nav can link to: the ones without an id.
func navRoutes(part *ir.Part) []string {
	var out []string
	for _, pg := range part.Pages {
		if !strings.Contains(pg.Route, "[") {
			out = append(out, pg.Route)
		}
	}
	return out
}
