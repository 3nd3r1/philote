package nextjs

import (
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

// callFiles renders the server action a calling component invokes. It fills
// the viewer's inputs on the server, so the client never sees or sends who
// the visitor is.
func callFiles(app *ir.App, backend *ir.Part, c *ir.Component) []target.File {
	if c.Calls == nil {
		return nil
	}
	call := c.Calls
	ep := endpoint(backend, call.Endpoint)
	used := map[string]bool{}
	v := actionCallView{Component: c.Name, Endpoint: call.Endpoint, Func: camel(call.Endpoint), Viewer: len(call.Viewer) > 0,
		Result: resultName(call), Value: outputType(ep, used)}
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

// passedType is what the component passes the action: a record's id, or a
// value of a scalar input, with enum values inline.
func passedType(app *ir.App, call *ir.Call) string {
	var fields []string
	for _, in := range call.Passed {
		t := "string"
		if in.Field != nil {
			t = scalarType(app, in.Field)
		}
		fields = append(fields, in.Name+": "+t)
	}
	return "{ " + strings.Join(fields, "; ") + " }"
}

func scalarType(app *ir.App, f *ir.Field) string {
	switch f.Kind {
	case ir.Int:
		return "number"
	case ir.Bool:
		return "boolean"
	case ir.EnumValue:
		return union(app.Enum(f.Enum).Values)
	default:
		return "string"
	}
}

// resultName is the type of what the action resolves to, like SwipeResult.
func resultName(call *ir.Call) string { return pascal(call.Endpoint) + "Result" }

// outputType is what an endpoint returns, in TypeScript. It's empty when the
// endpoint returns nothing.
func outputType(ep *ir.Endpoint, used map[string]bool) string {
	one := func(v *ir.Value) string {
		used[v.Entity] = true
		switch {
		case v.List:
			return v.Entity + "[]"
		case v.Optional:
			return v.Entity + " | null"
		}
		return v.Entity
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
