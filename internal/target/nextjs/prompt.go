package nextjs

import (
	"fmt"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
)

// themeRules tells the model how to style with the app's theme without
// knowing its values, so changing the look rewrites only the theme.
const themeRules = `The app uses shadcn/ui in the base-vega style with Tailwind CSS v4. The app's theme sets its colors, fonts and corner radius, and can change, so never assume a particular color, font or radius: use the theme's classes. Corners follow the theme with rounded-md, rounded-lg, rounded-xl and up.
Colors come only from the theme's variables: bg-background, text-foreground, bg-card, text-card-foreground, bg-muted, text-muted-foreground, bg-primary, text-primary-foreground, bg-secondary, text-secondary-foreground, bg-accent, border, ring. Headings can use font-heading.
`

// componentPrompt is everything the model is told about one component. It's
// built only from that component and the pages that use it, so editing
// another component or an unrelated entity doesn't change its key (D-019).
func componentPrompt(app *ir.App, part *ir.Part, c *ir.Component) string {
	var b strings.Builder
	kind := "server"
	if c.Calls != nil {
		kind = "client"
	}
	fmt.Fprintf(&b, "Write the body of one React %s component in a Next.js app that philote generated.\n%s\n", kind, themeRules)
	fmt.Fprintf(&b, "The component, %s, %s. %s\n", c.Name, subject(c), usage(part, c))
	names, extras := bodyParams(app, c)
	fmt.Fprintf(&b, "\nThe function whose body you write:\nfunction %sBody(%s): React.ReactNode\n\n", c.Name, bodyArgs(c, names, extras))
	if c.Param != "" {
		typ := propsType(app, c, "    ")
		if c.List {
			typ = "Array<" + typ + ">"
		}
		fmt.Fprintf(&b, "Its props, exactly. Nothing else is available:\ntype %sProps = {\n    %s: %s;\n};\n\n", c.Name, c.Param, typ)
		b.WriteString("Show every one of these, in this order of importance:\n")
		for _, s := range c.Shows {
			fmt.Fprintf(&b, "- %s.%s: %s\n", c.Param, s.Name, showHint(app, s))
		}
	}
	placed(&b, c)
	if c.Look != "" {
		fmt.Fprintf(&b, "\nHow it should look: %s\n", c.Look)
	}
	b.WriteString(`
Already imported. Use only these:
- Card, CardHeader, CardTitle, CardDescription, CardAction, CardContent, CardFooter: shadcn's card parts.
- Badge, with variant "default" | "secondary" | "destructive" | "outline" | "ghost" | "link".
- Separator, with orientation "horizontal" | "vertical".
- Icons: every lucide-react icon, as <Icons.Heart className="size-4" />.
- formatDate(iso: string): string, formatDateTime(iso: string): string, formatBool(value: boolean): string.
- cn(...classes) to join class names.
`)
	if c.LinksTo != "" || slices.ContainsFunc(c.Shows, func(s *ir.Show) bool { return s.Href != "" }) {
		b.WriteString("- Link from next/link.\n")
	}
	if c.Picks != "" {
		b.WriteString("- Pick({ id, className, children }): the button that picks one record.\n")
	}
	if c.Calls != nil {
		b.WriteString("- useState, useEffect, useRef and useTransition from react, and toast from sonner.\n")
	}
	b.WriteString(rules(c))
	return b.String()
}

// rules are what every body must keep to. A component that calls an
// endpoint runs in the browser and may move with a gesture; any other
// renders once on the server.
func rules(c *ir.Component) string {
	running := `- It renders on the server: no hooks, no state, no event handlers, no "use client".
`
	style := "No inline styles"
	if c.Calls != nil {
		running = `- It runs in the browser, already marked "use client": hooks, state and event handlers are fine. Use pointer events for gestures, so they work with a mouse and a finger.
- Never fetch or call anything else: act only through ` + camel(c.Calls.Endpoint) + `.
`
		style = "No inline styles, except a transform or opacity that follows a gesture"
	}
	return `
Rules:
- Write only the statements inside the function, ending with a return of JSX. No imports, no other functions.
` + running + `- Show only what the props hold. Don't invent data, placeholder text or images.
- Show each value once. Don't repeat a field in a second place or a second format.
- Words come only from the props, says and link.label. Don't write any other text, except a field's name as its label.
- Keep the card parts' own layout. CardAction already sits at the top right of its CardHeader.
- Style with Tailwind classes and the theme's variables. ` + style + `, no hard-coded colors.
- Reply with the statements in one ` + "```tsx" + ` block and nothing else.
`
}

func subject(c *ir.Component) string {
	switch {
	case c.Param == "":
		return "shows no record"
	case c.List:
		return "shows a list of " + c.Entity + " records"
	}
	return "shows one " + c.Entity
}

// placed tells the model about what the template hands the body besides
// its record: the link, the actions and the spec's words.
func placed(b *strings.Builder, c *ir.Component) {
	switch {
	case c.LinkLabel != "":
		b.WriteString("\nlink.href is another page and link.label is its button's text. Render <Link href={link.href}> with {link.label} in it as the main call to action, styled as a button.\n")
	case c.LinksTo != "":
		b.WriteString("\nlink.href is the record's own page. Make the whole card one link to it: a <Link href={link.href}> around the title, stretched over the card (the card relative, the link after:absolute after:inset-0). Anything else clickable sits above it with relative z-10.\n")
	}
	if len(c.Actions) > 0 {
		b.WriteString("\nactions holds the record's buttons. Render {actions} exactly once, where buttons belong, like the bottom of the card.\n")
	}
	if c.Picks != "" {
		fmt.Fprintf(b, "\nTapping a %[1]s makes it the visitor's own %[1]s. Wrap what shows each record in <Pick id={item.id} className=\"...\">...</Pick>, exactly one Pick per record: it renders a <button> with your className around its children, so style it as the tappable thing.\n", c.Entity)
	}
	if c.Calls != nil {
		callPrompt(b, c)
	}
	if len(c.Says) > 0 {
		b.WriteString("\nsays holds the spec's words, in order. Render each one exactly once, as {says[0]} and so on, never retyped:\n")
		for i, s := range c.Says {
			fmt.Fprintf(b, "- says[%d]: %s\n", i, jsString(s))
		}
	}
}

// usage says where pages put the component, so it can size itself: a card
// in a page's grid, or a full-width section.
func usage(part *ir.Part, c *ir.Component) string {
	var uses []string
	for _, pg := range part.Pages {
		for _, sec := range pg.Sections {
			switch {
			case sec.Component != c.Name:
			case sec.List:
				uses = append(uses, fmt.Sprintf("It's one card in a grid of cards on the %s page: return one Card that fills its cell (h-full), and the grid sets its width.", pg.Route))
			default:
				uses = append(uses, fmt.Sprintf("It's a full-width section of the %s page, under the app's navbar: it sets its own width (center its content in max-w-5xl with px-4 sm:px-6, or go full-bleed), padding and spacing, for phones and laptops.", pg.Route))
			}
		}
	}
	if len(uses) == 0 {
		return "No page uses it yet."
	}
	return strings.Join(uses, " ")
}

// propsType spells out the props' record type field by field, with enum
// values inline, so the prompt doesn't depend on the rest of types.ts.
func propsType(app *ir.App, c *ir.Component, indent string) string {
	ent := app.Entity(c.Entity)
	lines := []string{"id: string;"}
	for _, f := range ent.Fields {
		if shown(c, f.Name) || slices.Contains(c.LinkBy, f.Name) {
			lines = append(lines, fmt.Sprintf("%s%s: %s;", f.Name, optional(f), promptType(app, f)))
		}
	}
	for _, s := range c.Shows {
		if s.Relation == nil {
			continue
		}
		display := findField(app.Entity(s.Relation.Target), s.Display)
		t := fmt.Sprintf("{ id: string; %s%s: %s }", s.Display, optional(display), promptType(app, display))
		if s.Href != "" {
			t = fmt.Sprintf("{ id: string; %s%s: %s; href: string }", s.Display, optional(display), promptType(app, display))
		}
		if s.Relation.Many {
			t += "[]"
		}
		lines = append(lines, fmt.Sprintf("%s: %s;", s.Name, t))
	}
	inner := indent + "    "
	return "{\n" + inner + strings.Join(lines, "\n"+inner) + "\n" + indent + "}"
}

func showHint(app *ir.App, s *ir.Show) string {
	switch {
	case s.Relation != nil && s.Relation.Many && s.Href != "":
		return fmt.Sprintf("a list of %s records; show each one's %s as a <Link href={item.href}> to its page", s.Relation.Target, s.Display)
	case s.Relation != nil && s.Relation.Many:
		return fmt.Sprintf("a list of %s records; show each one's %s", s.Relation.Target, s.Display)
	case s.Relation != nil && s.Href != "":
		return fmt.Sprintf("a %s; show its %s as a <Link href={...href}> to its page", s.Relation.Target, s.Display)
	case s.Relation != nil:
		return fmt.Sprintf("a %s; show its %s", s.Relation.Target, s.Display)
	}
	switch s.Field.Kind {
	case ir.Date:
		return "a date as an ISO string; show it with formatDate"
	case ir.DateTime:
		return "a date and time as an ISO string; show it with formatDateTime"
	case ir.Bool:
		return "a boolean; show it with formatBool or as an icon"
	case ir.EnumValue:
		return "one of " + strings.Join(quoted(app.Enum(s.Field.Enum).Values), ", ")
	case ir.Slug:
		return "a URL-safe handle made from another field"
	}
	return promptType(app, s.Field)
}

func promptType(app *ir.App, f *ir.Field) string {
	if f.Kind == ir.EnumValue {
		return strings.Join(quoted(app.Enum(f.Enum).Values), " | ")
	}
	return tsType(f)
}

func optional(f *ir.Field) string {
	if f.Optional {
		return "?"
	}
	return ""
}

func shown(c *ir.Component, name string) bool {
	for _, s := range c.Shows {
		if s.Field != nil && s.Name == name {
			return true
		}
	}
	return false
}

func findField(e *ir.Entity, name string) *ir.Field {
	for _, f := range e.Fields {
		if f.Name == name {
			return f
		}
	}
	panic("nextjs: " + e.Name + " has no field " + name)
}
