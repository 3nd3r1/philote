package nextjs

import (
	"fmt"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
)

// componentPrompt is everything the model is told about one component. It's
// built only from that component and the pages that use it, so editing
// another component or an unrelated entity doesn't change its key (D-019).
func componentPrompt(app *ir.App, part *ir.Part, c *ir.Component) string {
	var b strings.Builder
	b.WriteString(`Write the body of one React server component in a Next.js app that philote generated.
The app uses shadcn/ui in the base-sera style (square corners, uppercase tracking-wide labels, compact type) with Tailwind CSS v4.
Colors come only from the theme's variables: bg-background, text-foreground, bg-card, text-card-foreground, bg-muted, text-muted-foreground, bg-primary, text-primary-foreground, bg-secondary, text-secondary-foreground, bg-accent, border, ring. Headings can use font-heading.

`)
	fmt.Fprintf(&b, "The component, %s, shows one %s. %s\n", c.Name, c.Entity, usage(part, c))
	if c.LinksTo != "" {
		b.WriteString("It's already wrapped in a link to the record's own page, so don't add links or anything else clickable.\n")
	}
	fmt.Fprintf(&b, "\nThe function whose body you write:\nfunction %sBody({ %s }: %sProps): React.ReactNode\n\n", c.Name, c.Param, c.Name)
	fmt.Fprintf(&b, "Its props, exactly. Nothing else is available:\ntype %sProps = {\n    %s: %s;\n};\n\n", c.Name, c.Param, propsType(app, c, "    "))
	b.WriteString("Show every one of these, in this order of importance:\n")
	for _, s := range c.Shows {
		fmt.Fprintf(&b, "- %s.%s: %s\n", c.Param, s.Name, showHint(app, s))
	}
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

Rules:
- Write only the statements inside the function, ending with a return of JSX. No imports, no other functions.
- It renders on the server: no hooks, no state, no event handlers, no "use client".
- Show only what the props hold. Don't invent data, placeholder text or images.
- Style with Tailwind classes and the theme's variables. No inline styles, no hard-coded colors.
- Reply with the statements in one ` + "```tsx" + ` block and nothing else.
`)
	return b.String()
}

// usage says where pages put the component, so it can size itself.
func usage(part *ir.Part, c *ir.Component) string {
	var uses []string
	for _, pg := range part.Pages {
		switch {
		case pg.Component != c.Name:
		case pg.ByID:
			uses = append(uses, fmt.Sprintf("It's the whole content of the %s page, in a centered column about 900px wide.", pg.Route))
		default:
			uses = append(uses, fmt.Sprintf("It's one item in a two-column grid of cards on the %s page.", pg.Route))
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
	case s.Relation != nil && s.Relation.Many:
		return fmt.Sprintf("a list of %s records; show each one's %s", s.Relation.Target, s.Display)
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
