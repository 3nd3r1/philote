package lower

import (
	"fmt"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
)

var kindNames = map[ir.Kind]string{
	ir.String: "string", ir.Slug: "slug", ir.Email: "email", ir.Markdown: "markdown",
	ir.Int: "int", ir.Bool: "bool", ir.Date: "date", ir.DateTime: "datetime",
}

// FieldType prints a field's type the way it's written in the source.
func FieldType(f *ir.Field) string {
	s := kindNames[f.Kind]
	if f.Kind == ir.EnumValue {
		s = f.Enum
	}
	var limits []string
	if f.Min != nil {
		limits = append(limits, fmt.Sprintf("min %d", *f.Min))
	}
	if f.Max != nil {
		limits = append(limits, fmt.Sprintf("max %d", *f.Max))
	}
	if len(limits) > 0 {
		s += "(" + strings.Join(limits, ", ") + ")"
	}
	if f.SlugFrom != "" {
		s += " from " + f.SlugFrom
	}
	return s
}

// Tree prints the app as a tree, down to depth levels below the root.
func Tree(app *ir.App, depth int) string {
	var b strings.Builder
	b.WriteString(app.Name + "\n")
	if depth < 1 {
		return b.String()
	}
	for _, p := range app.Parts {
		partTree(&b, p, depth)
	}
	fmt.Fprintf(&b, "  data (%d entities)\n", len(app.Entities))
	if depth >= 2 {
		for _, e := range app.Entities {
			fmt.Fprintf(&b, "    %s\n", e.Name)
		}
	}
	return b.String()
}

func partTree(b *strings.Builder, p *ir.Part, depth int) {
	kind := map[ir.PartKind]string{ir.Backend: "backend", ir.Frontend: "frontend"}[p.Kind]
	fmt.Fprintf(b, "  %s (%s, %s)", p.Name, kind, p.Target)
	if p.Uses != "" {
		b.WriteString(" uses " + p.Uses)
	}
	b.WriteString("\n")
	if depth < 2 {
		return
	}
	for _, r := range p.Resources {
		fmt.Fprintf(b, "    %s: %s\n", r.Name, r.Entity)
		if depth < 3 {
			continue
		}
		for _, ep := range p.Endpoints {
			if ep.Resource == r.Name {
				fmt.Fprintf(b, "      %s %s\n", ep.Method, ep.Route)
			}
		}
	}
	for _, ep := range p.Custom {
		fmt.Fprintf(b, "    %s (endpoint)\n", ep.Name)
		if depth >= 3 {
			fmt.Fprintf(b, "      %s %s\n      steps: %s\n", ep.Method, ep.Route, stepList(ep.Flow))
		}
	}
	viewTree(b, p, depth)
}

// viewTree prints a frontend's pages, then its components and what they show.
func viewTree(b *strings.Builder, p *ir.Part, depth int) {
	if p.Look != nil {
		fmt.Fprintf(b, "    theme: from its look\n")
	}
	for _, pg := range p.Pages {
		fmt.Fprintf(b, "    %s: %s\n", pg.Route, pageSummary(pg))
	}
	for _, c := range p.Components {
		componentTree(b, c, depth)
	}
}

// pageSummary is what a page loads and shows, or the form it submits.
func pageSummary(pg *ir.Page) string {
	if pg.Form != "" {
		return "form " + pg.Form + ", then " + pg.Then
	}
	var shows []string
	for _, s := range pg.Sections {
		show := s.Component
		if s.Arg {
			show += "(" + pg.Load + ")"
		}
		if s.List {
			show = "list of " + show
		}
		shows = append(shows, show)
	}
	out := strings.Join(shows, ", ")
	switch {
	case pg.ByID:
		out = "load " + pg.Load + " from " + pg.Resource + " by id, " + out
	case pg.Load != "":
		out = "load " + pg.Load + " from " + pg.Resource + ", " + out
	}
	return out
}

// componentTree prints one component and, at depth 3, what it shows and links to.
func componentTree(b *strings.Builder, c *ir.Component, depth int) {
	switch {
	case c.Param == "":
		fmt.Fprintf(b, "    %s\n", c.Name)
	case c.List:
		fmt.Fprintf(b, "    %s(%s: list of %s)\n", c.Name, c.Param, c.Entity)
	default:
		fmt.Fprintf(b, "    %s(%s: %s)\n", c.Name, c.Param, c.Entity)
	}
	if depth < 3 {
		return
	}
	var shows []string
	for _, s := range c.Shows {
		show := c.Param + "." + s.Name
		if s.Relation != nil && s.Relation.Many {
			show += " as chips"
		}
		shows = append(shows, show)
	}
	if len(shows) > 0 {
		fmt.Fprintf(b, "      shows: %s\n", strings.Join(shows, ", "))
	}
	if len(c.Says) > 0 {
		fmt.Fprintf(b, "      says: %s\n", strings.Join(c.Says, " / "))
	}
	if c.LinksTo != "" {
		fmt.Fprintf(b, "      links to %s\n", c.LinksTo)
	}
	for _, s := range c.Shows {
		if s.Href != "" {
			fmt.Fprintf(b, "      %s.%s links to %s\n", c.Param, s.Name, s.Href)
		}
	}
	for _, a := range c.Actions {
		fmt.Fprintf(b, "      %s\n", action(a))
	}
}

// action prints what an action does, like "deletes via horses.delete, then /horses".
func action(a *ir.Action) string {
	var out string
	switch a.Kind {
	case ir.Open:
		return "opens " + a.Route
	case ir.Remove:
		out = "deletes via " + a.Endpoint
	}
	if a.Then != "" {
		out += ", then " + a.Then
	}
	return out
}

// Show prints the endpoints at a path ("api" or "api/products") as spec,
// so a zoomed-out resource can be read at the level below.
func Show(app *ir.App, path string) (string, bool) {
	partName, resName, _ := strings.Cut(path, "/")
	for _, p := range app.Parts {
		if p.Name != partName {
			continue
		}
		var parts []string
		for _, ep := range p.Endpoints {
			if resName == "" || ep.Resource == resName || (ep.Flow != nil && ep.Name == resName) {
				parts = append(parts, endpointSpec(ep))
			}
		}
		return strings.Join(parts, "\n"), len(parts) > 0
	}
	return "", false
}

func endpointSpec(ep *ir.Endpoint) string {
	if ep.Flow != nil {
		return flowSpec(ep)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "endpoint %s {\n  route: %s %q\n", ep.Name, ep.Method, ep.Route)
	if ep.Input != nil {
		var fields []string
		for _, in := range ep.Input {
			if in.Field != nil {
				fields = append(fields, in.Name+": "+FieldType(in.Field))
			} else {
				fields = append(fields, in.Name+": "+in.Ref+".id")
			}
		}
		suffix := ""
		if ep.Partial {
			suffix = " (all optional)"
		}
		fmt.Fprintf(&b, "  input: { %s }%s\n", strings.Join(fields, ", "), suffix)
	}
	if ep.Output != "" {
		out := ep.Output
		if ep.List {
			out = "[" + out + "]"
		}
		fmt.Fprintf(&b, "  output: %s\n", out)
	}
	if len(ep.Errors) > 0 {
		fmt.Fprintf(&b, "  errors: %s\n", strings.Join(ep.Errors, ", "))
	}
	b.WriteString("}\n")
	return b.String()
}

func stepList(f *ir.Flow) string {
	var names []string
	for _, s := range f.Steps {
		names = append(names, s.Name)
	}
	return strings.Join(names, ", ")
}

func valueType(v *ir.Value) string {
	if v.Entity != "" {
		return v.Entity
	}
	return FieldType(v.Field)
}

// flowSpec prints a custom endpoint with the steps it runs.
func flowSpec(ep *ir.Endpoint) string {
	var b strings.Builder
	fmt.Fprintf(&b, "endpoint %s {\n  route: %s %q\n", ep.Name, ep.Method, ep.Route)
	if len(ep.Flow.Input) > 0 {
		var fields []string
		for _, v := range ep.Flow.Input {
			fields = append(fields, v.Name+": "+valueType(v))
		}
		fmt.Fprintf(&b, "  input: { %s }\n", strings.Join(fields, ", "))
	}
	if ep.Output != "" {
		fmt.Fprintf(&b, "  output: %s\n", ep.Output)
	}
	fmt.Fprintf(&b, "  errors: %s\n  steps: %s\n}\n", strings.Join(ep.Errors, ", "), stepList(ep.Flow))
	return b.String()
}
