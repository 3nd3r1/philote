package nextjs

import (
	"strings"
	"testing"
)

func fenced(theme string) string {
	return "@import \"tailwindcss\";\n\n" + themeStart + "\n" + theme + "\n" + themeEnd + "\n"
}

func TestDefaultThemePasses(t *testing.T) {
	if p := checkTheme(fenced(defaultTheme)); len(p) > 0 {
		t.Fatalf("the default theme fails its own check:\n%s", strings.Join(p, "\n"))
	}
}

func TestThemeProblems(t *testing.T) {
	cases := []struct{ name, old, new, want string }{
		{"missing color", "    --ring: oklch(0.6 0.12 265);\n}\n\n.dark", "}\n\n.dark", ":root: --ring is missing"},
		{"bad color", "--primary: oklch(0.45 0.16 265);", "--primary: #ff0000;", "--primary: use oklch(L C H)"},
		{"lightness over 1", "--primary: oklch(0.45 0.16 265);", "--primary: oklch(45 0.16 265);", "--primary: use oklch(L C H)"},
		{"unknown font", "--font-heading: var(--font-geist);", "--font-heading: var(--font-comic-sans);", "a font from the list"},
		{"radius too big", "--radius: 0.75rem;", "--radius: 4rem;", "from 0 to 2"},
		{"alpha on a color", "--card: oklch(1 0 0);", "--card: oklch(1 0 0 / 50%);", "only border and input"},
		{"unknown variable", "--radius: 0.75rem;", "--radius: 0.75rem;\n    --glow: oklch(1 0 0);", "--glow isn't a theme variable"},
		{"radius in dark", ".dark {\n", ".dark {\n    --radius: 1rem;\n", ".dark: --radius isn't a theme variable"},
		{"set twice", "--radius: 0.75rem;", "--radius: 0.75rem;\n    --radius: 1rem;", "--radius is set twice"},
		{"low contrast", "--muted-foreground: oklch(0.5 0.012 60);", "--muted-foreground: oklch(0.9 0.012 60);", "--muted-foreground on --background has contrast"},
		{"three rules", "\n\n.dark {", "\n\n.light {\n}\n\n.dark {", "exactly two rules"},
		{"stray text", "--radius: 0.75rem;", "radius is round", "expected `--name: value;`"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(defaultTheme, tc.old) {
				t.Fatalf("the default theme has no %q", tc.old)
			}
			p := strings.Join(checkTheme(fenced(strings.Replace(defaultTheme, tc.old, tc.new, 1))), "\n")
			if !strings.Contains(p, tc.want) {
				t.Errorf("no problem containing %q; got:\n%s", tc.want, p)
			}
		})
	}
}

// Each font a theme can name has to be loaded by the base's layout.
func TestLayoutLoadsEveryFont(t *testing.T) {
	layout, err := base.ReadFile("base/src/app/layout.tsx")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fonts {
		if !strings.Contains(string(layout), `variable: "--font-`+f.id+`"`) {
			t.Errorf("layout.tsx doesn't load %s as --font-%s", f.id, f.id)
		}
	}
}
