package nextjs

import (
	"fmt"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// navFile holds the nav: a wrapper the template owns, and its body.
const navFile = "src/components/common/layouts/main-navbar.tsx"

// navWrappers stick the nav to its place, so the model draws only what's
// inside and can't break the page's layout.
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
	Marker         string // the slot for the model's body; empty for the template's top bar
}

// nav renders the nav's items and the bar that draws them. Without a nav
// line, the bar is the template's own top bar over defaults.
func nav(app *ir.App, part *ir.Part, defaults []navItemView) ([]target.File, []target.Slot) {
	v := navView{Place: "top", Items: defaults}
	var slots []target.Slot
	if part.Nav != nil {
		v = navView{Place: part.Nav.Place, Marker: marker}
		for _, it := range part.Nav.Items {
			v.Items = append(v.Items, navItemView{jsString(it.Route), jsString(navLabel(app, part, it))})
		}
		slots = append(slots, target.Slot{Node: part.Name + "/nav", File: navFile, Marker: marker, Prompt: navPrompt(app, part)})
	}
	v.Wrapper = navWrappers[v.Place]
	return []target.File{
		render("nav.ts.tmpl", "src/lib/nav.ts", v),
		render("main-navbar.tsx.tmpl", navFile, v),
	}, slots
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

// navPlaces tell the model where its nav sits and what the wrapper already
// does there.
var navPlaces = map[string]string{
	"top":    "a bar across the top of every page. The wrapper already makes it sticky at the top and full width.",
	"bottom": "a tab bar at the bottom of the screen, like a phone app's, on every screen size. The wrapper already makes it sticky at the bottom and full width. Give it a solid background and a top border so pages scroll under it cleanly.",
	"side":   "a column on the left of every page, as tall as the screen. The wrapper already makes it sticky and full height. On phones keep it narrow, icons only with the label as aria-label; show the labels from sm up.",
}

// navPrompt is everything the model is told about the nav. It's built only
// from the nav line and the app's look, so changing a page or a component
// doesn't change its key.
func navPrompt(app *ir.App, part *ir.Part) string {
	var b strings.Builder
	b.WriteString("Write the body of the nav of a Next.js app that philote generated. It renders on every page.\n" + themeRules + "\n")
	if part.Look != nil {
		fmt.Fprintf(&b, "The app should look like this: %s\n\n", part.Look.Text)
	}
	fmt.Fprintf(&b, "The nav is %s\n", navPlaces[part.Nav.Place])
	b.WriteString(`
The function whose body you write:
function NavBody({ items, active, home, toggle }: {
    items: { href: string; label: string }[];
    active?: string;
    home: { href: string; label: string };
    toggle: React.ReactNode;
}): React.ReactNode

- items are the nav's pages, in order. Render every one with items.map as a <Link href={item.href}> holding {item.label} and an icon that fits it. They are:
`)
	for _, it := range part.Nav.Items {
		fmt.Fprintf(&b, "  - %s: %s\n", it.Route, navLabel(app, part, it))
	}
	b.WriteString(`- active is the href of the item for the current page, or undefined. Mark that item with aria-current="page" and a style that stands out.
- home is the app's name and its home page. Render <Link href={home.href}>{home.label}</Link> as the brand, where it fits the place.
- toggle is the light and dark switch. Render {toggle} exactly once.

Already imported. Use only these:
- Link from next/link.
- Icons: every lucide-react icon, as <Icons.Heart className="size-5" />.
- buttonVariants({ variant, size }): shadcn's button classes, with variant "default" | "outline" | "secondary" | "ghost" | "link" and size "default" | "sm" | "lg" | "icon".
- cn(...classes) to join class names.

Rules:
- Write only the statements inside the function, ending with a return of JSX. No imports, no other functions.
- No hooks, no state, no event handlers: everything it needs is in its props.
- Words come only from the items' labels and home.label. Don't write any other text.
- It must work on a 390px phone and on a laptop, with no sideways scrolling of the page and every item in reach.
- Style with Tailwind classes and the theme's variables. No inline styles, no hard-coded colors.
- Reply with the statements in one ` + "```tsx" + ` block and nothing else.
`)
	return b.String()
}
