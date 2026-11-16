package nextjs

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	declaration = regexp.MustCompile(`^--([a-z-]+):\s*(.+);$`)
	oklch       = regexp.MustCompile(`^oklch\((\d*\.?\d+) (\d*\.?\d+) (\d*\.?\d+)( / \d{1,3}%)?\)$`)
	radius      = regexp.MustCompile(`^(\d*\.?\d+)rem$`)
	fontVar     = regexp.MustCompile(`^var\(--font-([a-z-]+)\)$`)
)

// contrasts are the pairs whose text has to be readable, as foreground,
// background and minimum WCAG ratio. Primary may go to 3, the bar for large
// text and UI parts, so a brand color doesn't have to turn dark.
var contrasts = []struct {
	fg, bg string
	min    float64
}{
	{"foreground", "background", 4.5},
	{"muted-foreground", "background", 4.5},
	{"card-foreground", "card", 4.5},
	{"popover-foreground", "popover", 4.5},
	{"secondary-foreground", "secondary", 4.5},
	{"accent-foreground", "accent", 4.5},
	{"primary-foreground", "primary", 3},
}

// themeRule is one of the theme's two rules, as variable name to value.
type themeRule struct {
	selector string
	line     int // where the rule starts, for messages
	values   map[string]string
}

// checkTheme finds the theme fenced in globals.css and returns what's wrong
// with it, as "line N: ..." like tsc's problems.
func checkTheme(css string) []string {
	lines := strings.Split(css, "\n")
	start, end := slices.Index(lines, themeStart), slices.Index(lines, themeEnd)
	if start < 0 || end < start {
		panic("nextjs: globals.css has no theme fence")
	}
	rules, problems := parseTheme(lines[start+1:end], start+2)
	if len(problems) > 0 {
		return problems
	}
	if len(rules) != 2 || rules[0].selector != ":root" || rules[1].selector != ".dark" {
		return []string{fmt.Sprintf("line %d: the theme is exactly two rules, :root and then .dark", start+2)}
	}
	extra := []string{"radius", "font-sans", "font-heading"}
	problems = append(problems, checkRule(rules[0], extra)...)
	return append(problems, checkRule(rules[1], nil)...)
}

// parseTheme reads rules of one declaration per line. first is the file's
// line number of lines[0].
func parseTheme(lines []string, first int) ([]themeRule, []string) {
	var rules []themeRule
	var problems []string
	var cur *themeRule
	for i, raw := range lines {
		l, n := strings.TrimSpace(raw), first+i
		switch {
		case l == "":
		case cur == nil && strings.HasSuffix(l, "{"):
			cur = &themeRule{strings.TrimSpace(strings.TrimSuffix(l, "{")), n, map[string]string{}}
		case cur != nil && l == "}":
			rules, cur = append(rules, *cur), nil
		case cur != nil && declaration.MatchString(l):
			m := declaration.FindStringSubmatch(l)
			if _, dup := cur.values[m[1]]; dup {
				problems = append(problems, fmt.Sprintf("line %d: --%s is set twice", n, m[1]))
			}
			cur.values[m[1]] = m[2]
		default:
			problems = append(problems, fmt.Sprintf("line %d: expected `--name: value;`, a rule's `{` or `}`, found %q", n, l))
		}
	}
	if cur != nil {
		problems = append(problems, fmt.Sprintf("line %d: %s has no closing }", cur.line, cur.selector))
	}
	return rules, problems
}

// checkRule checks that a rule sets every color and the extra variables,
// nothing else, each in its format, and that its text is readable.
func checkRule(r themeRule, extra []string) []string {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf("line %d: %s: ", r.line, r.selector)+fmt.Sprintf(format, args...))
	}
	want := slices.Clone(extra)
	for _, c := range colors {
		want = append(want, c.name)
	}
	for _, name := range sortedKeys(setOf(r.values)) {
		if !slices.Contains(want, name) {
			add("--%s isn't a theme variable here", name)
		}
	}
	for _, name := range want {
		v, ok := r.values[name]
		if !ok {
			add("--%s is missing", name)
			continue
		}
		if msg := checkValue(name, v); msg != "" {
			add("--%s: %s", name, msg)
		}
	}
	if len(problems) > 0 {
		return problems
	}
	for _, c := range contrasts {
		if ratio := contrast(r.values[c.fg], r.values[c.bg]); ratio < c.min {
			add("--%s on --%s has contrast %.1f; it needs %.1f", c.fg, c.bg, ratio, c.min)
		}
	}
	return problems
}

func checkValue(name, v string) string {
	switch name {
	case "radius":
		m := radius.FindStringSubmatch(v)
		if m == nil || parseFloat(m[1]) > 2 {
			return "use a rem value from 0 to 2, like 0.75rem"
		}
	case "font-sans", "font-heading":
		m := fontVar.FindStringSubmatch(v)
		if m == nil || !slices.ContainsFunc(fonts, func(f font) bool { return f.id == m[1] }) {
			return "use var(--font-<name>) with a font from the list"
		}
	default:
		m := oklch.FindStringSubmatch(v)
		if m == nil || parseFloat(m[1]) > 1 {
			return "use oklch(L C H) with L from 0 to 1"
		}
		if m[4] != "" && name != "border" && name != "input" {
			return "only border and input may have an alpha"
		}
	}
	return ""
}

// contrast is the WCAG contrast ratio of two oklch colors without alpha.
func contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	return (max(la, lb) + 0.05) / (min(la, lb) + 0.05)
}

// luminance converts oklch to linear sRGB, clamped to the gamut the way a
// browser shows it, and weighs the channels as WCAG does.
func luminance(color string) float64 {
	m := oklch.FindStringSubmatch(color)
	l, c, h := parseFloat(m[1]), parseFloat(m[2]), parseFloat(m[3])*math.Pi/180
	a, b := c*math.Cos(h), c*math.Sin(h)
	cube := func(x float64) float64 { return x * x * x }
	lc := cube(l + 0.3963377774*a + 0.2158037573*b)
	mc := cube(l - 0.1055613458*a - 0.0638541728*b)
	sc := cube(l - 0.0894841775*a - 1.291485548*b)
	clamp := func(x float64) float64 { return min(1, max(0, x)) }
	r := clamp(4.0767416621*lc - 3.3077115913*mc + 0.2309699292*sc)
	g := clamp(-1.2684380046*lc + 2.6097574011*mc - 0.3413193965*sc)
	bl := clamp(-0.0041960863*lc - 0.7034186147*mc + 1.707614701*sc)
	return 0.2126*r + 0.7152*g + 0.0722*bl
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("nextjs: " + s + " matched a number pattern but isn't a number")
	}
	return f
}

func setOf(m map[string]string) map[string]bool {
	out := map[string]bool{}
	for k := range m {
		out[k] = true
	}
	return out
}
