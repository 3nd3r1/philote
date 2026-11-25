package nextjs

import (
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// navFile holds the templated navigation bar.
const navFile = "src/components/common/layouts/main-navbar.tsx"

// navWrappers stick the nav to its place in the shell.
var navWrappers = map[string]string{
	"top":    "sticky top-0 z-40 w-full",
	"bottom": "sticky bottom-0 z-40 w-full",
	"side":   "sticky top-0 z-40 h-svh shrink-0",
}

// navItemView is one nav item; both are JavaScript string literals.
type navItemView struct{ Href, Label string }

type navView struct {
	Place, Wrapper string
	Items          []navItemView
}

// nav renders the nav's items and the bar that draws them for top, bottom
// or side placement, without calling a model.
func nav(app *ir.App, part *ir.Part, defaults []navItemView) []target.File {
	v := navView{Place: "top", Items: defaults}
	if part.Nav != nil {
		v = navView{Place: part.Nav.Place}
		for _, it := range part.Nav.Items {
			v.Items = append(v.Items, navItemView{jsString(it.Route), jsString(navLabel(app, part, it))})
		}
	}
	v.Wrapper = navWrappers[v.Place]
	return []target.File{
		render("nav.ts.tmpl", "src/lib/nav.ts", v),
		render("main-navbar.tsx.tmpl", navFile, v),
	}
}

// navLabel is an item's text: the spec's, or its page's title.
func navLabel(app *ir.App, part *ir.Part, it *ir.NavItem) string {
	if it.Label != "" {
		return it.Label
	}
	for _, pg := range part.Pages {
		if pg.Route == it.Route {
			return pageTitle(app, pg)
		}
	}
	panic("nextjs: the nav links to " + it.Route + ", which isn't a page")
}
