package hono

import (
	"fmt"
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

func newTypesView(app *ir.App) typesView {
	var v typesView
	for _, e := range app.Enums {
		quoted := strings.Split(strings.Trim(list(e.Values), "[]"), ", ")
		v.Enums = append(v.Enums, struct{ Name, Union string }{e.Name, strings.Join(quoted, " | ")})
	}
	for _, e := range app.Entities {
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

// typeNames lists every enum and entity type, for imports.
func typeNames(app *ir.App) string {
	var names []string
	for _, e := range app.Enums {
		names = append(names, e.Name)
	}
	for _, e := range app.Entities {
		names = append(names, e.Name)
	}
	return strings.Join(names, ", ")
}

type endpointView struct {
	Name, Func, Method, Route, Types string
	Steps                            []mountView
	Rules                            []ruleView
	Lines                            []string
}

func newEndpointView(app *ir.App, ep *ir.Endpoint) endpointView {
	v := endpointView{
		Name:   ep.Name,
		Func:   "route" + pascal(ep.Name),
		Method: strings.ToLower(ep.Method),
		Route:  ep.Route,
		Types:  typeNames(app),
	}
	for _, in := range ep.Flow.Input {
		if in.Entity != "" {
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
			if s.Result.Entity == ep.Output && ep.Output != "" {
				result = s.Result.Name
			}
		}
		v.Lines = append(v.Lines, call)
	}
	if result != "" {
		v.Lines = append(v.Lines, fmt.Sprintf("return ctx.json(%s);", result))
	} else {
		v.Lines = append(v.Lines, "return ctx.body(null, 204);")
	}
	return v
}

type stepView struct {
	Name, Func, Types, Fails, Params, Returns, Marker, Intent string
}

func newStepView(app *ir.App, s *ir.Step) stepView {
	v := stepView{Name: s.Name, Func: camel(s.Name), Types: typeNames(app), Returns: "void", Marker: marker, Intent: s.Intent}
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
	return v
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
	b.WriteString(`Data access, already imported. Tables are named after entities and hold plain rows:
  all(table: string): Row[]
  find(table: string, id: string): Row | undefined
  insert(table: string, values: Record<string, unknown>): Row  // assigns the id
  update(table: string, id: string, values: Record<string, unknown>): Row | undefined
  remove(table: string, id: string): boolean
Row is { id: string; [field: string]: unknown }. Cast rows to the types below, like find("Horse", id) as Horse | undefined.
A relation is stored as <name>_id on the row that belongs to the other. Dates and times are ISO 8601 strings.

Types, already imported:
`)
	b.WriteString(types)
	b.WriteString(`
Rules:
- Write only the statements inside the function: no imports, no signature, no other functions.
- Use only the data access functions and types above. No other modules, no network, no files.
- Reply with the statements in one ` + "```ts" + ` block and nothing else.
`)
	return b.String()
}
