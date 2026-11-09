package hono

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
)

// resourceView is everything resource.ts.tmpl prints, already formatted as TypeScript.
type resourceView struct {
	Name   string
	Func   string
	File   string
	Path   string
	Table  string
	Ops    string
	Rules  []ruleView
	Unique string
	Slugs  string
}

type ruleView struct{ Name, Rule string }

var ruleKinds = map[ir.Kind]string{
	ir.String: "string", ir.Slug: "slug", ir.Email: "email", ir.Markdown: "markdown",
	ir.Int: "int", ir.Bool: "bool", ir.Date: "date", ir.DateTime: "datetime", ir.EnumValue: "enum",
}

func newResourceView(app *ir.App, part *ir.Part, r *ir.Resource) resourceView {
	e := app.Entity(r.Entity)
	v := resourceView{
		Name:  r.Name,
		Func:  "mount" + pascal(r.Name),
		File:  r.Name,
		Path:  "/" + r.Name,
		Table: e.Name,
	}
	var ops []string
	for _, op := range r.Ops {
		ops = append(ops, op.String())
	}
	v.Ops = list(ops)
	for _, in := range inputsOf(part, r) {
		v.Rules = append(v.Rules, ruleView{in.Name, rule(app, in)})
	}
	var unique, slugs []string
	for _, f := range e.Fields {
		if f.Unique {
			unique = append(unique, f.Name)
		}
		if f.SlugFrom != "" {
			slugs = append(slugs, fmt.Sprintf("%s: %q", f.Name, f.SlugFrom))
		}
	}
	v.Unique = list(unique)
	v.Slugs = "{}"
	if len(slugs) > 0 {
		v.Slugs = "{ " + strings.Join(slugs, ", ") + " }"
	}
	return v
}

// inputsOf returns the body fields of the resource's create or update endpoint.
func inputsOf(part *ir.Part, r *ir.Resource) []*ir.Input {
	for _, ep := range part.Endpoints {
		if ep.Resource == r.Name && ep.Input != nil {
			return ep.Input
		}
	}
	return nil
}

func rule(app *ir.App, in *ir.Input) string {
	if in.Field == nil {
		return refRule(in.Ref)
	}
	return fieldRule(app, in.Field)
}

func refRule(entity string) string {
	return fmt.Sprintf("{ kind: \"ref\", table: %q }", entity)
}

func fieldRule(app *ir.App, f *ir.Field) string {
	parts := []string{fmt.Sprintf("kind: %q", ruleKinds[f.Kind])}
	if f.Min != nil {
		parts = append(parts, "min: "+strconv.Itoa(*f.Min))
	}
	if f.Max != nil {
		parts = append(parts, "max: "+strconv.Itoa(*f.Max))
	}
	if f.Kind == ir.EnumValue {
		parts = append(parts, "values: "+list(app.Enum(f.Enum).Values))
	}
	if f.Optional {
		parts = append(parts, "optional: true")
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// list prints a TypeScript array of string literals.
func list(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = strconv.Quote(s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// camel turns record_swipe into recordSwipe.
func camel(s string) string {
	p := pascal(s)
	return strings.ToLower(p[:1]) + p[1:]
}

// pascal turns order_lines into OrderLines.
func pascal(s string) string {
	var b strings.Builder
	for _, w := range strings.Split(s, "_") {
		if w != "" {
			b.WriteString(strings.ToUpper(w[:1]) + w[1:])
		}
	}
	return b.String()
}
