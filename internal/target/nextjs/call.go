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
}

// callFiles renders the server action a calling component's body is
// handed. It fills the viewer's inputs on the server, so the body never
// sees or sends who the visitor is.
func callFiles(app *ir.App, backend *ir.Part, c *ir.Component) []target.File {
	if c.Calls == nil {
		return nil
	}
	call := c.Calls
	v := actionCallView{Component: c.Name, Endpoint: call.Endpoint, Func: camel(call.Endpoint), Viewer: len(call.Viewer) > 0}
	if len(call.Passed) > 0 {
		v.Params = "input: " + passedType(app, call)
	}
	var fields []string
	for _, in := range endpoint(backend, call.Endpoint).Flow.Input {
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
		return "() => Promise<string | undefined>"
	}
	return fmt.Sprintf("(input: %s) => Promise<string | undefined>", passedType(app, call))
}

// callPrompt tells the model what the action does and what to pass it.
func callPrompt(b *strings.Builder, c *ir.Component) {
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
	b.WriteString(" It resolves to an error message to show with toast.error(message), or undefined when it worked. Call it only when the visitor acts, never while rendering, and keep what's on screen in step with it.\n")
}
