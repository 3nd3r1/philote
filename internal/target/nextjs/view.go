package nextjs

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
)

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

type fieldView struct {
	Name, Type string
	Optional   bool
}

type typeView struct {
	Name   string
	Fields []fieldView
}

type typesView struct {
	Enums    []struct{ Name, Union string }
	Entities []typeView
}

// newTypesView gives the shapes the backend returns: every field, plus the
// foreign key of each relation an entity belongs to.
func newTypesView(app *ir.App) typesView {
	var v typesView
	for _, e := range app.Enums {
		v.Enums = append(v.Enums, struct{ Name, Union string }{e.Name, strings.Join(quoted(e.Values), " | ")})
	}
	for _, e := range app.Entities {
		t := typeView{Name: e.Name}
		for _, f := range e.Fields {
			t.Fields = append(t.Fields, fieldView{f.Name, tsType(f), f.Optional})
		}
		for _, r := range e.Relations {
			if !r.Many {
				t.Fields = append(t.Fields, fieldView{r.FK, "string", false})
			}
		}
		v.Entities = append(v.Entities, t)
	}
	return v
}

type apiView struct {
	Backend   string
	Imports   string
	Inputs    []typeView
	Resources []struct {
		Name  string
		Calls []callView
	}
	Endpoints []callView
}

// callView is one client function and the request it sends.
type callView struct {
	Name, Params, Value, Errors, Method, Path, ErrorList string
	Input                                                bool
}

// newAPIView builds the client for a backend's endpoints. Resources become
// objects of calls, like horses.get; custom endpoints become functions.
func newAPIView(app *ir.App, backend *ir.Part) (apiView, error) {
	v := apiView{Backend: backend.Name}
	used := map[string]bool{}
	for _, r := range backend.Resources {
		res := struct {
			Name  string
			Calls []callView
		}{Name: camel(r.Name)}
		for _, ep := range backend.Endpoints {
			if ep.Resource != r.Name {
				continue
			}
			c := newCall(ep, camel(strings.TrimPrefix(ep.Name, r.Name+".")))
			if ep.Input != nil {
				input := addInput(&v, "New"+ep.Entity, resourceInput(ep.Input, used))
				if ep.Partial {
					input = "Partial<" + input + ">"
				}
				c.Params = strings.TrimPrefix(c.Params+", input: "+input, ", ")
			}
			res.Calls = append(res.Calls, c)
			used[ep.Output] = true
		}
		v.Resources = append(v.Resources, res)
	}
	for _, ep := range backend.Custom {
		c := newCall(ep, camel(ep.Name))
		if len(ep.Flow.Input) > 0 {
			c.Params = "input: " + addInput(&v, pascal(ep.Name)+"Input", flowInput(ep.Flow.Input, used))
		}
		v.Endpoints = append(v.Endpoints, c)
		used[ep.Output] = true
	}
	v.Imports = imports(app, used)
	return v, checkNames(app, v)
}

func newCall(ep *ir.Endpoint, name string) callView {
	c := callView{
		Name:      name,
		Value:     "void",
		Errors:    "never",
		Method:    ep.Method,
		Path:      strconv.Quote(ep.Route),
		ErrorList: "[" + strings.Join(quoted(ep.Errors), ", ") + "]",
		Input:     ep.Input != nil || (ep.Flow != nil && len(ep.Flow.Input) > 0),
	}
	if ep.Output != "" {
		c.Value = ep.Output
		if ep.List {
			c.Value += "[]"
		}
	}
	if len(ep.Errors) > 0 {
		c.Errors = strings.Join(quoted(ep.Errors), " | ")
	}
	if strings.Contains(ep.Route, "[id]") {
		c.Params = "id: string"
		c.Path = "`" + strings.Replace(ep.Route, "[id]", "${encodeURIComponent(id)}", 1) + "`"
	}
	return c
}

// addInput declares an input type once, since create and update share it.
func addInput(v *apiView, name string, fields []fieldView) string {
	if !slices.ContainsFunc(v.Inputs, func(t typeView) bool { return t.Name == name }) {
		v.Inputs = append(v.Inputs, typeView{name, fields})
	}
	return name
}

// resourceInput is a resource's request body. Foreign keys are sent as ids.
func resourceInput(inputs []*ir.Input, used map[string]bool) []fieldView {
	var out []fieldView
	for _, in := range inputs {
		if in.Field == nil {
			out = append(out, fieldView{in.Name, "string", false})
			continue
		}
		out = append(out, fieldView{in.Name, tsType(in.Field), in.Field.Optional})
		used[in.Field.Enum] = true
	}
	return out
}

// flowInput is a custom endpoint's request body. Records are sent as ids.
func flowInput(values []*ir.Value, used map[string]bool) []fieldView {
	var out []fieldView
	for _, val := range values {
		if val.Field == nil {
			out = append(out, fieldView{val.Name, "string", false})
			continue
		}
		out = append(out, fieldView{val.Name, tsType(val.Field), val.Field.Optional})
		used[val.Field.Enum] = true
	}
	return out
}

// imports lists the used types in the order types.ts declares them.
func imports(app *ir.App, used map[string]bool) string {
	var names []string
	for _, e := range app.Enums {
		if used[e.Name] {
			names = append(names, e.Name)
		}
	}
	for _, e := range app.Entities {
		if used[e.Name] {
			names = append(names, e.Name)
		}
	}
	return strings.Join(names, ", ")
}

// checkNames refuses a client that would declare one name twice, such as
// an input type NewHorse next to an entity called NewHorse.
func checkNames(app *ir.App, v apiView) error {
	seen := map[string]bool{"call": true, "Issue": true, "Result": true}
	for _, e := range app.Enums {
		seen[e.Name] = true
	}
	for _, e := range app.Entities {
		seen[e.Name] = true
	}
	var names []string
	for _, t := range v.Inputs {
		names = append(names, t.Name)
	}
	for _, r := range v.Resources {
		names = append(names, r.Name)
		calls := map[string]bool{}
		for _, c := range r.Calls {
			if calls[c.Name] {
				return fmt.Errorf("the client for %s would declare %s.%s twice; rename the relation", v.Backend, r.Name, c.Name)
			}
			calls[c.Name] = true
		}
	}
	for _, c := range v.Endpoints {
		names = append(names, c.Name)
	}
	for _, n := range names {
		if seen[n] {
			return fmt.Errorf("the client for %s would declare %s twice; rename the endpoint, resource or entity behind it", v.Backend, n)
		}
		seen[n] = true
	}
	return nil
}

func quoted(items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = strconv.Quote(s)
	}
	return out
}

// camel turns order_lines into orderLines.
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
