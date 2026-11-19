package hono

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
)

// marker is the line in a step file that the model's body replaces.
const marker = "  // philote:body"

func tsType(f *ir.Field) string {
	switch f.Kind {
	case ir.Int:
		return "number"
	case ir.Bool:
		return "boolean"
	case ir.EnumValue:
		return f.Enum
	}
	return "string"
}

func valueType(v *ir.Value) string {
	if v.List {
		return v.Entity + "[]"
	}
	if v.Optional {
		return v.Entity + " | null"
	}
	if v.Entity != "" {
		return v.Entity
	}
	return tsType(v.Field)
}

type typesView struct {
	Enums    []struct{ Name, Union string }
	Entities []entityTypeView
}

type entityTypeView struct {
	Name   string
	Fields []struct {
		Name, Type string
		Optional   bool
	}
}

func newTypesView(enums []*ir.Enum, entities []*ir.Entity) typesView {
	var v typesView
	for _, e := range enums {
		quoted := strings.Split(strings.Trim(list(e.Values), "[]"), ", ")
		v.Enums = append(v.Enums, struct{ Name, Union string }{e.Name, strings.Join(quoted, " | ")})
	}
	for _, e := range entities {
		ev := entityTypeView{Name: e.Name}
		add := func(name, typ string, optional bool) {
			ev.Fields = append(ev.Fields, struct {
				Name, Type string
				Optional   bool
			}{name, typ, optional})
		}
		for _, f := range e.Fields {
			add(f.Name, tsType(f), f.Optional)
		}
		for _, r := range e.Relations {
			if !r.Many {
				add(r.FK, "string", false)
			}
		}
		v.Entities = append(v.Entities, ev)
	}
	return v
}

// typeNames lists enum and entity types, for imports.
func typeNames(enums []*ir.Enum, entities []*ir.Entity) string {
	var names []string
	for _, e := range enums {
		names = append(names, e.Name)
	}
	for _, e := range entities {
		names = append(names, e.Name)
	}
	return strings.Join(names, ", ")
}

// stepTypes is what a step reaches: the entities of its parameters, result
// and tables, and the enums they use, in spec order. Nothing else goes in its
// prompt, so an edit elsewhere in the data leaves its key alone (D-022).
func stepTypes(app *ir.App, s *ir.Step) ([]*ir.Enum, []*ir.Entity) {
	names := map[string]bool{}
	for _, v := range append(slices.Clone(s.Params), s.Result) {
		switch {
		case v == nil:
		case v.Entity != "":
			names[v.Entity] = true
		case v.Field.Kind == ir.EnumValue:
			names[v.Field.Enum] = true
		}
	}
	for _, t := range slices.Concat(s.Reads, s.Writes) {
		names[t] = true
	}
	var entities []*ir.Entity
	for _, e := range app.Entities {
		if names[e.Name] {
			entities = append(entities, e)
			for _, f := range e.Fields {
				if f.Kind == ir.EnumValue {
					names[f.Enum] = true
				}
			}
		}
	}
	var enums []*ir.Enum
	for _, e := range app.Enums {
		if names[e.Name] {
			enums = append(enums, e)
		}
	}
	return enums, entities
}

type endpointView struct {
	Name, Func, Method, Route, Types string
	Find                             bool // the handler finds records by id
	Steps                            []mountView
	Rules                            []ruleView
	Lines                            []string
}

// routeParam matches a route param like [horse].
var routeParam = regexp.MustCompile(`\[(\w+)\]`)

func newEndpointView(app *ir.App, ep *ir.Endpoint) endpointView {
	v := endpointView{
		Name:   ep.Name,
		Func:   "route" + pascal(ep.Name),
		Method: strings.ToLower(ep.Method),
		Route:  routeParam.ReplaceAllString(ep.Route, ":$1"),
		Types:  typeNames(app.Enums, app.Entities),
	}
	// A record in the route that doesn't exist is not_found, like a
	// resource's get; one in the body is invalid_input.
	for _, in := range ep.Flow.Params {
		v.Find = true
		v.Lines = append(v.Lines,
			fmt.Sprintf("const %s = find(%q, ctx.req.param(%q)) as %s | undefined;", in.Name, in.Entity, in.Name, in.Entity),
			fmt.Sprintf("if (!%s) {", in.Name),
			`  return ctx.json({ error: "not_found" }, 404);`,
			"}")
	}
	for _, in := range ep.Flow.Input {
		if in.Entity != "" {
			v.Find = true
			v.Rules = append(v.Rules, ruleView{in.Name, refRule(in.Entity)})
			v.Lines = append(v.Lines, fmt.Sprintf("const %s = find(%q, checked.values.%s as string) as %s;", in.Name, in.Entity, in.Name, in.Entity))
		} else {
			v.Rules = append(v.Rules, ruleView{in.Name, fieldRule(app, in.Field)})
			v.Lines = append(v.Lines, fmt.Sprintf("const %s = checked.values.%s as %s;", in.Name, in.Name, tsType(in.Field)))
		}
	}
	result := ""
	for _, s := range ep.Flow.Steps {
		v.Steps = append(v.Steps, mountView{camel(s.Name), s.Name})
		var args []string
		for _, p := range s.Params {
			args = append(args, p.Name)
		}
		call := fmt.Sprintf("await %s(%s);", camel(s.Name), strings.Join(args, ", "))
		if s.Result != nil {
			call = fmt.Sprintf("const %s = %s", s.Result.Name, call)
			if s.Result.Entity == ep.Output && s.Result.List == ep.List && ep.Output != "" {
				result = s.Result.Name
			}
		}
		v.Lines = append(v.Lines, call)
	}
	switch {
	case len(ep.Outputs) > 0:
		names := make([]string, len(ep.Outputs))
		for i, o := range ep.Outputs {
			names[i] = o.Name
		}
		v.Lines = append(v.Lines, fmt.Sprintf("return ctx.json({ %s });", strings.Join(names, ", ")))
	case result != "":
		v.Lines = append(v.Lines, fmt.Sprintf("return ctx.json(%s);", result))
	default:
		v.Lines = append(v.Lines, "return ctx.body(null, 204);")
	}
	return v
}

type stepView struct {
	Name, Func, Types, Fails, Params, Returns, Marker, Intent string
	Readable, Writable                                        string // unions of table names; empty for none
}

func newStepView(app *ir.App, s *ir.Step) stepView {
	v := stepView{Name: s.Name, Func: camel(s.Name), Types: typeNames(stepTypes(app, s)), Returns: "void", Marker: marker, Intent: s.Intent}
	var params, fails []string
	for _, p := range s.Params {
		params = append(params, p.Name+": "+valueType(p))
	}
	for _, f := range s.Fails {
		fails = append(fails, fmt.Sprintf("%q", f))
	}
	v.Params = strings.Join(params, ", ")
	v.Fails = strings.Join(fails, " | ")
	if s.Result != nil {
		v.Returns = valueType(s.Result)
	}
	v.Readable = union(slices.Concat(s.Reads, s.Writes))
	v.Writable = union(s.Writes)
	return v
}

// union is a TypeScript union of string literals.
func union(names []string) string {
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = strconv.Quote(n)
	}
	return strings.Join(quoted, " | ")
}

// prompt tells the model everything it needs to write one step body. It's
// built only from the spec, so the same spec always gives the same prompt.
func prompt(v stepView, types string) string {
	var b strings.Builder
	b.WriteString("Write the body of one TypeScript function in a Hono backend that philote generated.\n\n")
	fmt.Fprintf(&b, "The function:\nexport async function %s(%s): Promise<%s>\n\n", v.Func, v.Params, v.Returns)
	fmt.Fprintf(&b, "What it does: %s\n\n", v.Intent)
	if v.Fails != "" {
		fmt.Fprintf(&b, "It may fail with %s. Fail by calling fail(\"<error>\"), which throws. Use no other errors.\n\n", strings.ReplaceAll(v.Fails, " | ", ", "))
	}
	if v.Readable == "" {
		b.WriteString("It has no data access: work only from its parameters.\n\n")
	} else {
		b.WriteString("Data access, already in scope. Tables are named after entities and hold plain rows. Only these tables are allowed:\n")
		fmt.Fprintf(&b, "  all(table: %s): Row[]\n", v.Readable)
		fmt.Fprintf(&b, "  find(table: %s, id: string): Row | undefined\n", v.Readable)
		if v.Writable != "" {
			fmt.Fprintf(&b, "  insert(table: %s, values: Record<string, unknown>): Row  // assigns the id\n", v.Writable)
			fmt.Fprintf(&b, "  update(table: %s, id: string, values: Record<string, unknown>): Row | undefined\n", v.Writable)
			fmt.Fprintf(&b, "  remove(table: %s, id: string): boolean\n", v.Writable)
		}
		b.WriteString(`Row is { id: string; [field: string]: unknown }. Cast rows to the types below, like find("Horse", id) as Horse | undefined.
A relation is stored as <name>_id on the row that belongs to the other. Dates and times are ISO 8601 strings.

`)
	}
	if v.Types != "" {
		b.WriteString("Types, already imported:\n")
		b.WriteString(types)
		b.WriteString("\n")
	}
	b.WriteString(`Rules:
- Write only the statements inside the function: no imports, no signature, no other functions.
- Use only the data access functions and types above. No other modules, no network, no files.
- Reply with the statements in one ` + "```ts" + ` block and nothing else.
`)
	return b.String()
}
