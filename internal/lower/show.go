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
		kind := map[ir.PartKind]string{ir.Backend: "backend", ir.Frontend: "frontend"}[p.Kind]
		fmt.Fprintf(&b, "  %s (%s, %s)\n", p.Name, kind, p.Target)
		if depth < 2 {
			continue
		}
		for _, r := range p.Resources {
			fmt.Fprintf(&b, "    %s: %s\n", r.Name, r.Entity)
			if depth < 3 {
				continue
			}
			for _, ep := range p.Endpoints {
				if ep.Resource == r.Name {
					fmt.Fprintf(&b, "      %s %s\n", ep.Method, ep.Route)
				}
			}
		}
	}
	fmt.Fprintf(&b, "  data (%d entities)\n", len(app.Entities))
	if depth >= 2 {
		for _, e := range app.Entities {
			fmt.Fprintf(&b, "    %s\n", e.Name)
		}
	}
	return b.String()
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
			if resName == "" || ep.Resource == resName {
				parts = append(parts, endpointSpec(ep))
			}
		}
		return strings.Join(parts, "\n"), len(parts) > 0
	}
	return "", false
}

func endpointSpec(ep *ir.Endpoint) string {
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
