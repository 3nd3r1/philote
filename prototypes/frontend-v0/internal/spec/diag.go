package spec

import (
	"fmt"
	"sort"
	"strings"
)

type Pos struct{ Line, Col int }

func (p Pos) String() string { return fmt.Sprintf("%d:%d", p.Line, p.Col) }

type Severity int

const (
	SevError Severity = iota
	SevWarning
)

type Diag struct {
	Sev  Severity
	Pos  Pos
	Msg  string
	Hint string
}

type Diags struct{ List []Diag }

func (d *Diags) Errorf(p Pos, format string, args ...any) {
	d.List = append(d.List, Diag{SevError, p, fmt.Sprintf(format, args...), ""})
}

func (d *Diags) Warnf(p Pos, format string, args ...any) {
	d.List = append(d.List, Diag{SevWarning, p, fmt.Sprintf(format, args...), ""})
}

// Hint attaches a hint to the most recent diagnostic.
func (d *Diags) Hint(format string, args ...any) {
	if n := len(d.List); n > 0 {
		d.List[n-1].Hint = fmt.Sprintf(format, args...)
	}
}

func (d *Diags) HasErrors() bool {
	for _, x := range d.List {
		if x.Sev == SevError {
			return true
		}
	}
	return false
}

func (d *Diags) Sorted() []Diag {
	out := append([]Diag(nil), d.List...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].Pos, out[j].Pos
		return a.Line < b.Line || (a.Line == b.Line && a.Col < b.Col)
	})
	return out
}

// Render prints diagnostics compiler-style, with the source line and a caret.
func Render(path, src string, list []Diag) string {
	lines := strings.Split(src, "\n")
	var b strings.Builder
	for _, x := range list {
		sev := "error"
		if x.Sev == SevWarning {
			sev = "warning"
		}
		fmt.Fprintf(&b, "%s:%d:%d: %s: %s\n", path, x.Pos.Line, x.Pos.Col, sev, x.Msg)
		if x.Pos.Line >= 1 && x.Pos.Line <= len(lines) {
			gutter := fmt.Sprintf("%4d", x.Pos.Line)
			fmt.Fprintf(&b, "%s | %s\n", gutter, strings.ReplaceAll(lines[x.Pos.Line-1], "\t", " "))
			fmt.Fprintf(&b, "%s | %s^\n", strings.Repeat(" ", len(gutter)), strings.Repeat(" ", max(0, x.Pos.Col-1)))
		}
		if x.Hint != "" {
			fmt.Fprintf(&b, "     = hint: %s\n", x.Hint)
		}
	}
	return b.String()
}

func levenshtein(a, b string) int {
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

// hintNames adds "did you mean X?" if a candidate is close, otherwise lists the candidates.
func (d *Diags) hintNames(name string, cands []string, what string) {
	best, bestD := "", 3
	for _, c := range cands {
		if dist := levenshtein(name, c); dist < bestD && dist < len(name) {
			best, bestD = c, dist
		}
	}
	switch {
	case best != "":
		d.Hint("did you mean '%s'?", best)
	case len(cands) > 0:
		d.Hint("%s: %s", what, strings.Join(cands, ", "))
	}
}
