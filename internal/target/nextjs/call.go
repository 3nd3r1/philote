package nextjs

import (
	"fmt"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// actionCallView is the server action behind a component's calls. Params and
// Input are TypeScript: the action's parameter list and the client call's
// argument.
type actionCallView struct {
	Component, Endpoint, Func string
	Viewer                    bool // the action fills inputs with the viewer
	Params, Input             string
	Result, Value, Types      string // the result type's name, what a call that worked holds, and imports from @/lib/types
}

// callFiles renders the server action a calling component's body is
// handed. It fills the viewer's inputs on the server, so the body never
// sees or sends who the visitor is.
func callFiles(app *ir.App, backend *ir.Part, c *ir.Component) []target.File {
	if c.Calls == nil {
		return nil
	}
	call := c.Calls
	ep := endpoint(backend, call.Endpoint)
	used := map[string]bool{}
	v := actionCallView{Component: c.Name, Endpoint: call.Endpoint, Func: camel(call.Endpoint), Viewer: len(call.Viewer) > 0,
		Result: resultName(call), Value: outputType(ep, used, func(entity string) string { return entity })}
	v.Types = imports(app, used)
	if len(call.Passed) > 0 {
		v.Params = "input: " + passedType(app, call)
	}
	var fields []string
	for _, in := range ep.Flow.Input {
		if slices.Contains(call.Viewer, in.Name) {
			fields = append(fields, in.Name+": viewer.id")
		} else {
			fields = append(fields, in.Name+": input."+in.Name)
		}
	}
	if len(fields) > 0 {
		v.Input = "{ " + strings.Join(fields, ", ") + " }"
	}
	return []target.File{render("call.ts.tmpl", "src/lib/actions/"+kebab(c.Name)+"-call.ts", v)}
}

// passedType is what the body passes the action: a record's id, or a value
// of a scalar input, with enum values inline.
func passedType(app *ir.App, call *ir.Call) string {
	var fields []string
	for _, in := range call.Passed {
		t := "string"
		if in.Field != nil {
			t = promptType(app, in.Field)
		}
		fields = append(fields, in.Name+": "+t)
	}
	return "{ " + strings.Join(fields, "; ") + " }"
}

// actionType is the type of the action the body receives.
func actionType(app *ir.App, call *ir.Call) string {
	if len(call.Passed) == 0 {
		return "() => Promise<" + resultName(call) + ">"
	}
	return fmt.Sprintf("(input: %s) => Promise<%s>", passedType(app, call), resultName(call))
}

// resultName is the type of what the action resolves to, like SwipeResult.
func resultName(call *ir.Call) string { return pascal(call.Endpoint) + "Result" }

// outputType is what an endpoint returns, in TypeScript, with each record
// spelled by record: its name in code, or its fields in a prompt. It's
// empty when the endpoint returns nothing.
func outputType(ep *ir.Endpoint, used map[string]bool, record func(entity string) string) string {
	one := func(v *ir.Value) string {
		used[v.Entity] = true
		switch {
		case v.List:
			return record(v.Entity) + "[]"
		case v.Optional:
			return record(v.Entity) + " | null"
		}
		return record(v.Entity)
	}
	switch {
	case len(ep.Outputs) > 0:
		fields := make([]string, len(ep.Outputs))
		for i, o := range ep.Outputs {
			fields[i] = o.Name + ": " + one(o)
		}
		return "{ " + strings.Join(fields, "; ") + " }"
	case ep.Output != "":
		return one(&ir.Value{Entity: ep.Output, List: ep.List})
	}
	return ""
}

// recordType spells out an entity's record as the backend sends it, with
// a belongs-to relation as its id, so the prompt doesn't need types.ts.
func recordType(app *ir.App, entity string) string {
	ent := app.Entity(entity)
	fields := []string{"id: string"}
	for _, f := range ent.Fields {
		fields = append(fields, fmt.Sprintf("%s%s: %s", f.Name, optional(f), promptType(app, f)))
	}
	for _, r := range ent.Relations {
		if !r.Many {
			fields = append(fields, r.FK+": string")
		}
	}
	return "{ " + strings.Join(fields, "; ") + " }"
}

// callPrompt tells the model what the action does, what to pass it, and
// what it resolves to.
func callPrompt(b *strings.Builder, app *ir.App, backend *ir.Part, c *ir.Component) {
	call := c.Calls
	name := camel(call.Endpoint)
	fmt.Fprintf(b, "\n%s sends the %s endpoint", name, call.Endpoint)
	if len(call.Viewer) > 0 {
		fmt.Fprintf(b, " as the visitor, who fills %s on the server", strings.Join(call.Viewer, " and "))
	}
	b.WriteString(". ")
	if len(call.Passed) == 0 {
		fmt.Fprintf(b, "Call it as await %s().", name)
	} else {
		var args, hints []string
		for _, in := range call.Passed {
			args = append(args, in.Name)
			switch {
			case in.Entity != "" && c.List:
				hints = append(hints, fmt.Sprintf("%s is the id of the one of %s's records it acts on", in.Name, c.Param))
			case in.Entity != "":
				hints = append(hints, fmt.Sprintf("%s is %s.id", in.Name, c.Param))
			default:
				hints = append(hints, fmt.Sprintf("%s comes from what the visitor does", in.Name))
			}
		}
		fmt.Fprintf(b, "Call it as await %s({ %s }): %s.", name, strings.Join(args, ", "), strings.Join(hints, "; "))
	}
	result := resultName(call)
	value := outputType(endpoint(backend, call.Endpoint), map[string]bool{}, func(entity string) string { return recordType(app, entity) })
	if value == "" {
		fmt.Fprintf(b, " It resolves to a %[1]s:\ntype %[1]s = { ok: true } | { ok: false; error: string };\n", result)
	} else {
		fmt.Fprintf(b, " It resolves to a %[1]s, with what the call made in value:\ntype %[1]s = { ok: true; value: %[2]s } | { ok: false; error: string };\n", result, value)
	}
	b.WriteString("Show error with toast.error(error). When it worked, show what value holds if it's worth a moment, like something the call made. Call it only when the visitor acts, never while rendering, and keep what's on screen in step with it.\n")
}
