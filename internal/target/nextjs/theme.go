package nextjs

import (
	"fmt"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// themeMarker is the line in globals.css that the model's theme replaces.
const themeMarker = "/* philote:theme */"

// themeStart and themeEnd fence the theme in globals.css, so Verify can
// find it whether a template or the model wrote it.
const (
	themeStart = "/* theme: start */"
	themeEnd   = "/* theme: end */"
)

// font is one Google font the base's layout loads, as --font-<id>.
type font struct{ id, feel string }

// fonts are the only fonts a theme can pick. layout.tsx loads each of them
// without preloading, so only the ones a theme uses are downloaded.
var fonts = []font{
	{"geist", "neutral, modern sans"},
	{"inter", "plain, highly legible sans"},
	{"dm-sans", "friendly geometric sans"},
	{"manrope", "clean, slightly technical sans"},
	{"nunito", "rounded, soft, playful sans"},
	{"space-grotesk", "quirky, techy grotesque"},
	{"bricolage-grotesque", "expressive, characterful display sans"},
	{"fraunces", "warm, soft serif with character"},
	{"playfair-display", "elegant, high-contrast serif"},
	{"dm-serif-display", "bold, classic display serif, headings only"},
	{"instrument-serif", "refined, editorial serif, headings only"},
}

// colors are the theme's color variables, set for both light and dark.
var colors = []struct{ name, use string }{
	{"background", "the page"},
	{"foreground", "text on the page"},
	{"card", "cards"},
	{"card-foreground", "text on cards"},
	{"popover", "menus and popovers"},
	{"popover-foreground", "text on popovers"},
	{"primary", "the main accent: primary buttons, highlights"},
	{"primary-foreground", "text on primary"},
	{"secondary", "secondary buttons and badges"},
	{"secondary-foreground", "text on secondary"},
	{"muted", "quiet backgrounds"},
	{"muted-foreground", "quiet text, like captions and dates"},
	{"accent", "hover and selected backgrounds"},
	{"accent-foreground", "text on accent"},
	{"destructive", "delete and error"},
	{"border", "borders"},
	{"input", "form field borders"},
	{"ring", "focus rings"},
}

// defaultTheme is the look of a frontend without a look line: calm, light
// or dark with the system, and readable.
const defaultTheme = `:root {
    --radius: 0.75rem;
    --font-sans: var(--font-geist);
    --font-heading: var(--font-geist);
    --background: oklch(0.99 0.004 85);
    --foreground: oklch(0.2 0.01 60);
    --card: oklch(1 0 0);
    --card-foreground: oklch(0.2 0.01 60);
    --popover: oklch(1 0 0);
    --popover-foreground: oklch(0.2 0.01 60);
    --primary: oklch(0.45 0.16 265);
    --primary-foreground: oklch(0.98 0.01 265);
    --secondary: oklch(0.95 0.01 85);
    --secondary-foreground: oklch(0.25 0.01 60);
    --muted: oklch(0.955 0.006 85);
    --muted-foreground: oklch(0.5 0.012 60);
    --accent: oklch(0.94 0.03 265);
    --accent-foreground: oklch(0.3 0.1 265);
    --destructive: oklch(0.577 0.245 27.325);
    --border: oklch(0.9 0.008 85);
    --input: oklch(0.88 0.008 85);
    --ring: oklch(0.6 0.12 265);
}

.dark {
    --background: oklch(0.17 0.01 265);
    --foreground: oklch(0.96 0.005 85);
    --card: oklch(0.21 0.012 265);
    --card-foreground: oklch(0.96 0.005 85);
    --popover: oklch(0.21 0.012 265);
    --popover-foreground: oklch(0.96 0.005 85);
    --primary: oklch(0.72 0.14 265);
    --primary-foreground: oklch(0.18 0.03 265);
    --secondary: oklch(0.27 0.012 265);
    --secondary-foreground: oklch(0.96 0.005 85);
    --muted: oklch(0.25 0.01 265);
    --muted-foreground: oklch(0.72 0.01 85);
    --accent: oklch(0.3 0.04 265);
    --accent-foreground: oklch(0.96 0.005 85);
    --destructive: oklch(0.704 0.191 22.216);
    --border: oklch(1 0 0 / 10%);
    --input: oklch(1 0 0 / 15%);
    --ring: oklch(0.6 0.12 265);
}`

// globals renders globals.css with the default theme, or with a slot for
// the model's theme when the frontend has a look line.
func globals(part *ir.Part) (target.File, []target.Slot) {
	if part.Look == nil {
		return render("globals.css.tmpl", "src/app/globals.css", defaultTheme), nil
	}
	slot := target.Slot{
		Node:   part.Name + "/theme",
		File:   "src/app/globals.css",
		Marker: themeMarker,
		Prompt: themePrompt(part.Look.Text),
	}
	return render("globals.css.tmpl", slot.File, themeMarker), []target.Slot{slot}
}

// themePrompt is built only from the look line and the target, so changing
// a page or a component never rewrites the theme, and changing the look
// rewrites nothing else.
func themePrompt(look string) string {
	var b strings.Builder
	b.WriteString(`Write the theme of a web app that philote generated: the CSS variables that shadcn/ui and Tailwind CSS v4 read, for light and dark mode.

How the app should look: ` + look + `

Reply with one ` + "```css" + ` block holding exactly two rules and nothing else: ` + "`:root`" + ` for light mode, then ` + "`.dark`" + ` for dark mode. One declaration per line, no comments.

` + "`:root`" + ` sets these three, and every color:
- --radius: the corner radius, in rem from 0 to 2. 0 is square, 0.5rem is soft, 1rem and up is round and friendly.
- --font-sans: the body font, as var(--font-<name>).
- --font-heading: the font for headings, as var(--font-<name>).

The fonts, by name:
`)
	for _, f := range fonts {
		fmt.Fprintf(&b, "- %s: %s\n", f.id, f.feel)
	}
	b.WriteString("\n`.dark` sets only the colors. Every color, in both rules:\n")
	for _, c := range colors {
		fmt.Fprintf(&b, "- --%s: %s\n", c.name, c.use)
	}
	b.WriteString(`
Rules:
- Every color is oklch(L C H), with L from 0 to 1, like oklch(0.62 0.19 25). Only border and input may add an alpha, like oklch(1 0 0 / 12%).
- Text must be readable: each foreground against its own background has a contrast ratio of at least 4.5 (primary-foreground on primary at least 3), and so does muted-foreground on background.
- Light and dark are the same design: same hues and character, with lightness flipped.
- Make it look like a polished product someone designed on purpose, true to the look above.

The format, with placeholder values:
` + "```css\n" + defaultTheme + "\n```\n")
	return b.String()
}
