package diag

import (
	"fmt"
	"slices"
	"strings"
)

// Pos is a place in a source file. Line and Col start at 1.
type Pos struct {
	File      string
	Line, Col int
}

func (p Pos) String() string { return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col) }

type Severity int

const (
	Error Severity = iota
	Warning
)

type Diag struct {
	Sev  Severity
	Pos  Pos
	Msg  string
	Hint string
}

// List collects diagnostics in the order they were found.
type List struct {
	Items []Diag
}

func (l *List) Errorf(p Pos, format string, args ...any) {
	l.Items = append(l.Items, Diag{Error, p, fmt.Sprintf(format, args...), ""})
}

func (l *List) Warnf(p Pos, format string, args ...any) {
	l.Items = append(l.Items, Diag{Warning, p, fmt.Sprintf(format, args...), ""})
}

// Hint attaches a hint to the most recent diagnostic.
func (l *List) Hint(format string, args ...any) {
	if n := len(l.Items); n > 0 {
		l.Items[n-1].Hint = fmt.Sprintf(format, args...)
	}
}

// HintNames hints at the closest candidate, or lists them all if none is close.
func (l *List) HintNames(name string, candidates []string, what string) {
	if best := closest(name, candidates); best != "" {
		l.Hint("did you mean '%s'?", best)
	} else if len(candidates) > 0 {
		l.Hint("%s: %s", what, strings.Join(candidates, ", "))
	}
}

func (l *List) HasErrors() bool {
	return slices.ContainsFunc(l.Items, func(d Diag) bool { return d.Sev == Error })
}

// Sorted returns the diagnostics ordered by file and position.
func (l *List) Sorted() []Diag {
	out := slices.Clone(l.Items)
	slices.SortStableFunc(out, func(a, b Diag) int {
		if c := strings.Compare(a.Pos.File, b.Pos.File); c != 0 {
			return c
		}
		if a.Pos.Line != b.Pos.Line {
			return a.Pos.Line - b.Pos.Line
		}
		return a.Pos.Col - b.Pos.Col
	})
	return out
}

// Render prints diagnostics compiler-style, with the source line and a caret.
// sources maps a file name to its text.
func Render(diags []Diag, sources map[string]string) string {
	var b strings.Builder
	for _, d := range diags {
		sev := "error"
		if d.Sev == Warning {
			sev = "warning"
		}
		fmt.Fprintf(&b, "%s: %s: %s\n", d.Pos, sev, d.Msg)
		lines := strings.Split(sources[d.Pos.File], "\n")
		if d.Pos.Line >= 1 && d.Pos.Line <= len(lines) {
			gutter := fmt.Sprintf("%4d", d.Pos.Line)
			fmt.Fprintf(&b, "%s | %s\n", gutter, strings.ReplaceAll(lines[d.Pos.Line-1], "\t", " "))
			fmt.Fprintf(&b, "%s | %s^\n", strings.Repeat(" ", len(gutter)), strings.Repeat(" ", max(0, d.Pos.Col-1)))
		}
		if d.Hint != "" {
			fmt.Fprintf(&b, "     = hint: %s\n", d.Hint)
		}
	}
	return b.String()
}

func closest(name string, candidates []string) string {
	best, bestDist := "", 3
	for _, c := range candidates {
		if d := distance(name, c); d < bestDist && d < len(name) {
			best, bestDist = c, d
		}
	}
	return best
}

// distance is the Levenshtein edit distance.
func distance(a, b string) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}
