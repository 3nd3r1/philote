package nextjs

import (
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// themeStart and themeEnd fence the theme in globals.css, so Verify can
// find and contrast-check it.
const (
	themeStart = "/* theme: start */"
	themeEnd   = "/* theme: end */"
)

// font is one Google font the base's layout loads, as --font-<id>.
type font struct{ id, feel string }

// fonts are the fonts layout.tsx loads without preloading, so only the ones
// a theme uses are downloaded.
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

// builtInThemes are the fixed, contrast-verified CSS token sets a frontend
// picks with `theme: <name>` (decision-7).
var builtInThemes = map[string]string{
	"plain": `:root {
    --mode: system;
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
}`,

	"warm": `:root {
    --mode: light;
    --radius: 1rem;
    --font-sans: var(--font-nunito);
    --font-heading: var(--font-fraunces);
    --background: oklch(0.985 0.012 75);
    --foreground: oklch(0.22 0.02 35);
    --card: oklch(1 0.005 75);
    --card-foreground: oklch(0.22 0.02 35);
    --popover: oklch(1 0.005 75);
    --popover-foreground: oklch(0.22 0.02 35);
    --primary: oklch(0.56 0.19 28);
    --primary-foreground: oklch(0.99 0.005 75);
    --secondary: oklch(0.94 0.02 65);
    --secondary-foreground: oklch(0.26 0.03 35);
    --muted: oklch(0.95 0.014 70);
    --muted-foreground: oklch(0.48 0.025 40);
    --accent: oklch(0.93 0.04 35);
    --accent-foreground: oklch(0.32 0.1 28);
    --destructive: oklch(0.55 0.22 25);
    --border: oklch(0.89 0.02 65);
    --input: oklch(0.87 0.02 65);
    --ring: oklch(0.6 0.16 28);
}

.dark {
    --background: oklch(0.18 0.015 35);
    --foreground: oklch(0.96 0.01 75);
    --card: oklch(0.22 0.018 35);
    --card-foreground: oklch(0.96 0.01 75);
    --popover: oklch(0.22 0.018 35);
    --popover-foreground: oklch(0.96 0.01 75);
    --primary: oklch(0.74 0.15 28);
    --primary-foreground: oklch(0.18 0.03 28);
    --secondary: oklch(0.28 0.02 35);
    --secondary-foreground: oklch(0.96 0.01 75);
    --muted: oklch(0.26 0.015 35);
    --muted-foreground: oklch(0.73 0.015 70);
    --accent: oklch(0.32 0.05 28);
    --accent-foreground: oklch(0.96 0.01 75);
    --destructive: oklch(0.7 0.19 22);
    --border: oklch(1 0 0 / 12%);
    --input: oklch(1 0 0 / 16%);
    --ring: oklch(0.74 0.15 28);
}`,

	"editorial": `:root {
    --mode: light;
    --radius: 0.375rem;
    --font-sans: var(--font-dm-sans);
    --font-heading: var(--font-instrument-serif);
    --background: oklch(0.975 0.008 90);
    --foreground: oklch(0.21 0.015 110);
    --card: oklch(0.995 0.004 90);
    --card-foreground: oklch(0.21 0.015 110);
    --popover: oklch(0.995 0.004 90);
    --popover-foreground: oklch(0.21 0.015 110);
    --primary: oklch(0.38 0.08 135);
    --primary-foreground: oklch(0.98 0.006 90);
    --secondary: oklch(0.93 0.012 95);
    --secondary-foreground: oklch(0.24 0.02 115);
    --muted: oklch(0.94 0.01 92);
    --muted-foreground: oklch(0.47 0.018 105);
    --accent: oklch(0.91 0.025 130);
    --accent-foreground: oklch(0.26 0.06 135);
    --destructive: oklch(0.54 0.2 25);
    --border: oklch(0.88 0.014 95);
    --input: oklch(0.86 0.014 95);
    --ring: oklch(0.45 0.08 135);
}

.dark {
    --background: oklch(0.16 0.012 120);
    --foreground: oklch(0.95 0.008 90);
    --card: oklch(0.2 0.014 120);
    --card-foreground: oklch(0.95 0.008 90);
    --popover: oklch(0.2 0.014 120);
    --popover-foreground: oklch(0.95 0.008 90);
    --primary: oklch(0.72 0.1 135);
    --primary-foreground: oklch(0.16 0.02 135);
    --secondary: oklch(0.26 0.016 120);
    --secondary-foreground: oklch(0.95 0.008 90);
    --muted: oklch(0.24 0.014 120);
    --muted-foreground: oklch(0.72 0.014 95);
    --accent: oklch(0.29 0.035 135);
    --accent-foreground: oklch(0.95 0.008 90);
    --destructive: oklch(0.7 0.18 24);
    --border: oklch(1 0 0 / 12%);
    --input: oklch(1 0 0 / 16%);
    --ring: oklch(0.72 0.1 135);
}`,

	"slate": `:root {
    --mode: system;
    --radius: 0.5rem;
    --font-sans: var(--font-manrope);
    --font-heading: var(--font-space-grotesk);
    --background: oklch(0.985 0.005 245);
    --foreground: oklch(0.19 0.02 255);
    --card: oklch(1 0 0);
    --card-foreground: oklch(0.19 0.02 255);
    --popover: oklch(1 0 0);
    --popover-foreground: oklch(0.19 0.02 255);
    --primary: oklch(0.44 0.15 250);
    --primary-foreground: oklch(0.99 0.005 250);
    --secondary: oklch(0.94 0.012 245);
    --secondary-foreground: oklch(0.24 0.025 255);
    --muted: oklch(0.95 0.008 245);
    --muted-foreground: oklch(0.48 0.02 250);
    --accent: oklch(0.92 0.03 250);
    --accent-foreground: oklch(0.28 0.09 250);
    --destructive: oklch(0.56 0.22 25);
    --border: oklch(0.89 0.012 245);
    --input: oklch(0.87 0.012 245);
    --ring: oklch(0.55 0.14 250);
}

.dark {
    --background: oklch(0.15 0.015 255);
    --foreground: oklch(0.96 0.006 245);
    --card: oklch(0.19 0.018 255);
    --card-foreground: oklch(0.96 0.006 245);
    --popover: oklch(0.19 0.018 255);
    --popover-foreground: oklch(0.96 0.006 245);
    --primary: oklch(0.73 0.13 245);
    --primary-foreground: oklch(0.15 0.025 255);
    --secondary: oklch(0.25 0.02 255);
    --secondary-foreground: oklch(0.96 0.006 245);
    --muted: oklch(0.23 0.016 255);
    --muted-foreground: oklch(0.72 0.015 245);
    --accent: oklch(0.28 0.04 250);
    --accent-foreground: oklch(0.96 0.006 245);
    --destructive: oklch(0.7 0.19 22);
    --border: oklch(1 0 0 / 12%);
    --input: oklch(1 0 0 / 16%);
    --ring: oklch(0.73 0.13 245);
}`,
}

// globals renders globals.css with the frontend's fixed theme (plain by
// default).
func globals(part *ir.Part) target.File {
	name := part.Theme
	if name == "" {
		name = "plain"
	}
	css, ok := builtInThemes[name]
	if !ok {
		panic("nextjs: unknown theme " + name)
	}
	return render("globals.css.tmpl", "src/app/globals.css", css)
}
