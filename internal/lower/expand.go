package lower

import "github.com/3nd3r1/philote/internal/ir"

// shape is how a standard operation maps to HTTP.
type shape struct {
	method  string
	onItem  bool // route ends in /[id]
	input   bool
	partial bool
	output  bool
	list    bool
	errors  []string
}

var shapes = map[ir.Op]shape{
	ir.List:   {method: "GET", output: true, list: true},
	ir.Get:    {method: "GET", onItem: true, output: true, errors: []string{"not_found"}},
	ir.Create: {method: "POST", input: true, output: true, errors: []string{"invalid_input"}},
	ir.Update: {method: "PATCH", onItem: true, input: true, partial: true, output: true, errors: []string{"not_found", "invalid_input"}},
	ir.Delete: {method: "DELETE", onItem: true, errors: []string{"not_found"}},
}

// Expand fills each backend part's endpoints from its resources.
func Expand(app *ir.App) {
	for _, p := range app.Parts {
		p.Endpoints = nil
		for _, r := range p.Resources {
			p.Endpoints = append(p.Endpoints, endpoints(app, r)...)
		}
	}
}

func endpoints(app *ir.App, r *ir.Resource) []*ir.Endpoint {
	e := app.Entity(r.Entity)
	var out []*ir.Endpoint
	for _, op := range r.Ops {
		s := shapes[op]
		ep := &ir.Endpoint{
			Name:     r.Name + "." + op.String(),
			Method:   s.method,
			Route:    "/" + r.Name,
			Resource: r.Name,
			Op:       op,
			Entity:   e.Name,
			Partial:  s.partial,
			List:     s.list,
			Errors:   s.errors,
		}
		if s.onItem {
			ep.Route += "/[id]"
		}
		if s.input {
			ep.Input = inputs(e)
		}
		if s.output {
			ep.Output = e.Name
		}
		out = append(out, ep)
	}
	return out
}

// inputs are the fields a client sends: every field except slugs the server
// derives, plus a foreign key for each relation this entity belongs to.
func inputs(e *ir.Entity) []*ir.Input {
	var out []*ir.Input
	for _, f := range e.Fields {
		if f.Kind == ir.Slug && f.SlugFrom != "" {
			continue
		}
		out = append(out, &ir.Input{Name: f.Name, Field: f})
	}
	for _, r := range e.Relations {
		if !r.Many {
			out = append(out, &ir.Input{Name: r.FK, Ref: r.Target})
		}
	}
	return out
}
