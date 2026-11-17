package nextjs

import (
	"fmt"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// pageMarker is the line in a page file that the model's layout replaces.
const pageMarker = "    // philote:body"

// pageSlot is the hole for a page's layout. Its node is the part and the
// route, like web/horses/[id].
func pageSlot(app *ir.App, part *ir.Part, pg *ir.Page) target.Slot {
	return target.Slot{
		Node:   part.Name + pg.Route,
		File:   "src/app" + pg.Route + "/page.tsx",
		Marker: pageMarker,
		Prompt: pagePrompt(app, part, pg),
	}
}

// pagePrompt is built from the app's name and look and the page alone, not
// from the components it shows, so editing a component never rewrites a
// page (decision-2).
func pagePrompt(app *ir.App, part *ir.Part, pg *ir.Page) string {
	var b strings.Builder
	b.WriteString("Write the body of one page layout, a React server component, in a Next.js app that philote generated.\n" + themeRules + "\n")
	fmt.Fprintf(&b, "The app is called %s.", app.Name)
	if part.Look != nil {
		fmt.Fprintf(&b, " How it should look: %s", part.Look.Text)
	}
	b.WriteString("\n" + pageRole(pg) + "\n")
	b.WriteString(`
The function whose body you write:
function PageBody({ title, content }: PageBodyProps): React.ReactNode

Its props, exactly. Nothing else is available:
type PageBodyProps = {
    title: string;
    content: (className: string) => React.ReactNode;
};

The page is everything under the app's navbar, inside a full-width <main> that grows to fill the window. You set its width, padding and spacing, for phones and laptops. The navbar's content is a centered max-w-5xl column with px-4 sm:px-6: line the page's edges up with it, or center a narrower column for one record or a form.
`)
	b.WriteString(contentRole(pg))
	b.WriteString(`
Already imported. Use only these:
- Badge, with variant "default" | "secondary" | "destructive" | "outline" | "ghost" | "link".
- Separator, with orientation "horizontal" | "vertical".
- Icons: every lucide-react icon, as <Icons.Heart className="size-4" />.
- cn(...classes) to join class names.

Rules:
- Write only the statements inside the function, ending with a return of JSX. No imports, no other functions.
- It renders on the server: no hooks, no state, no event handlers, no "use client".
- Call content exactly once, with a string literal for its class.
- Besides the title you may add one short line saying what the page is for. Don't invent data, numbers, names, images or other text.
- Nothing clickable: the navbar and the cards hold the links and actions.
- Style with Tailwind classes and the theme's variables. No inline styles, no hard-coded colors.
- Reply with the statements in one ` + "```tsx" + ` block and nothing else.
`)
	return b.String()
}

// pageRole says what the page is for and what its title is.
func pageRole(pg *ir.Page) string {
	switch {
	case pg.Form != "" && pg.Resource == "":
		return fmt.Sprintf("The page %s is a form that submits %s. Its title, %q, is its heading.", pg.Route, words(pg.Form), formTitle(pg))
	case pg.Form != "" && pg.ByID:
		return fmt.Sprintf("The page %s is a form that edits one %s. Its title, %q, is its heading.", pg.Route, pg.Entity, formTitle(pg))
	case pg.Form != "":
		return fmt.Sprintf("The page %s is a form that makes a new %s. Its title, %q, is its heading.", pg.Route, pg.Entity, formTitle(pg))
	case pg.ByID:
		return fmt.Sprintf("The page %s shows one %s as a %s, which already names the record big and holds its actions. Its title, %q, is generic: show it small, like a label above the view, or leave it out.", pg.Route, pg.Entity, pg.Component, title(pg))
	}
	return fmt.Sprintf("The page %s lists every %s, each as a %s card. Its title, %q, is its heading.", pg.Route, pg.Entity, pg.Component, title(pg))
}

// contentRole says what content renders and what its class is for.
func contentRole(pg *ir.Page) string {
	switch {
	case pg.Form != "":
		return "content(className) renders the form's fields and its save button, in a <div> with that class. Size it for a form.\n"
	case pg.ByID:
		return "content(className) renders the view, in a <div> with that class. Size it for one record's view.\n"
	}
	return "content(className) renders every card in one <div> with that class, so the class decides how they're arranged: a grid, columns or a stack, sized for the cards and for phones. Each card fills its cell. When there are no records it renders a short message instead.\n"
}
