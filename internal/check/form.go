package check

import (
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// formPage checks a page that submits a form (D-020). It returns the
// entity of the record a save returns, which its `then` may link to.
func (c *checker) formPage(pg *ir.Page, backend *ir.Part, s *syntax.Page) (string, bool) {
	for _, k := range []struct {
		set  bool
		name string
	}{{s.Data != nil, "data"}, {s.Layout != nil, "layout"}, {s.Empty != nil, "empty"}} {
		if k.set {
			c.d.Errorf(s.Form.Pos, "a form page takes form and then, not %s", k.name)
			return "", false
		}
	}
	if s.FormOp == nil {
		return c.customForm(pg, backend, s)
	}
	res := findResource(backend, s.Form.Name)
	if res == nil {
		c.d.Errorf(s.Form.Pos, "%s has no resource named %s", backend.Name, s.Form.Name)
		c.d.HintNames(s.Form.Name, resourceNames(backend), "resources")
		return "", false
	}
	op := slices.Index(ir.OpNames[:], s.FormOp.Name)
	switch {
	case op != int(ir.Create) && op != int(ir.Update):
		c.d.Errorf(s.FormOp.Pos, "a form creates or updates; use %s.create or %s.update", res.Name, res.Name)
		return "", false
	case !slices.Contains(res.Ops, ir.Op(op)):
		c.d.Errorf(s.FormOp.Pos, "resource %s doesn't serve %s", res.Name, s.FormOp.Name)
		return "", false
	case op == int(ir.Update) && !slices.Contains(res.Ops, ir.Get):
		c.d.Errorf(s.FormOp.Pos, "an edit form loads the record first, but %s doesn't serve get", res.Name)
		return "", false
	}
	pg.Form, pg.Resource, pg.Entity = res.Name+"."+s.FormOp.Name, res.Name, res.Entity
	pg.ByID = op == int(ir.Update)
	if !c.formRoute(pg, s) || !c.formCopy(pg, s) {
		return "", false
	}
	ent := c.entities[res.Entity]
	for _, rel := range ent.Relations {
		if !rel.Many && !c.formRecord(backend, rel.Target, *s.Form) {
			return "", false
		}
	}
	return res.Entity, true
}

// customForm checks a form on a custom endpoint, like `form: swipe`.
func (c *checker) customForm(pg *ir.Page, backend *ir.Part, s *syntax.Page) (string, bool) {
	i := slices.IndexFunc(backend.Custom, func(e *ir.Endpoint) bool { return e.Name == s.Form.Name })
	if i < 0 {
		var names []string
		for _, e := range backend.Custom {
			names = append(names, e.Name)
		}
		c.d.Errorf(s.Form.Pos, "%s has no endpoint named %s", backend.Name, s.Form.Name)
		if findResource(backend, s.Form.Name) != nil {
			c.d.Hint("a resource's form is %s.create or %s.update", s.Form.Name, s.Form.Name)
		} else {
			c.d.HintNames(s.Form.Name, names, "endpoints")
		}
		return "", false
	}
	ep := backend.Custom[i]
	if len(ep.Flow.Input) == 0 {
		c.d.Errorf(s.Form.Pos, "endpoint %s takes no input, so it can't be a form", ep.Name)
		return "", false
	}
	pg.Form = ep.Name
	if !c.formRoute(pg, s) || !c.formCopy(pg, s) {
		return "", false
	}
	for _, v := range ep.Flow.Input {
		if v.Entity != "" && !c.formRecord(backend, v.Entity, *s.Form) {
			return "", false
		}
	}
	return ep.Output, true
}

func (c *checker) formRoute(pg *ir.Page, s *syntax.Page) bool {
	if s.By != nil && s.By.Name != "id" {
		c.d.Errorf(s.By.Pos, "forms load by id only")
		return false
	}
	hasID := strings.Contains(pg.Route, "[id]")
	switch {
	case pg.ByID && s.By == nil:
		c.d.Errorf(s.Form.Pos, "an edit form needs the record it edits")
		c.d.Hint("use `form: %s by id`", pg.Form)
	case !pg.ByID && s.By != nil:
		c.d.Errorf(s.By.Pos, "only an update form loads a record")
	case pg.ByID && !hasID:
		c.d.Errorf(s.Route.Pos, "an edit form needs [id] in its route")
	case !pg.ByID && hasID:
		c.d.Errorf(s.Route.Pos, "this form makes a new record, so its route can't have [id]")
	default:
		return true
	}
	return false
}

// formCopy checks that an edit form says what shows when its record is
// gone, like a page by id (D-006).
func (c *checker) formCopy(pg *ir.Page, s *syntax.Page) bool {
	switch {
	case s.Then == nil:
		c.d.Errorf(s.Word.Pos, "a form page must say where to go after saving")
		c.d.Hint(`add then: "<route>"`)
	case pg.ByID && s.NotFound == nil:
		c.d.Errorf(s.Word.Pos, "an edit form must say what shows when the record doesn't exist")
		c.d.Hint(`add not_found: "That %s doesn't exist."`, strings.ToLower(pg.Entity))
	case !pg.ByID && s.NotFound != nil:
		c.d.Errorf(s.NotFound.Pos, "only an edit form has not_found")
	case s.NotFound != nil && len(s.NotFound.Value) > maxCopy:
		c.d.Errorf(s.NotFound.Pos, "UI text is capped at %d characters", maxCopy)
	default:
		if s.NotFound != nil {
			pg.NotFound = s.NotFound.Value
		}
		return true
	}
	return false
}

// formRecord checks that a form can offer records of entity to pick from:
// it needs their display field and a resource that lists them.
func (c *checker) formRecord(backend *ir.Part, entity string, at syntax.Ident) bool {
	if c.entities[entity].Display == "" {
		c.d.Errorf(at.Pos, "this form picks a %s, but %s has no display field to show in the list", entity, entity)
		c.d.Hint("add `display: <field>` to %s, like `display: name`", entity)
		return false
	}
	return c.serving(backend, entity, ir.List, "lists", at) != ""
}

// then checks where a form goes after a save. Its [id] is the saved
// record's, so the page there must load that entity by id.
func (c *checker) then(part *ir.Part, pg *ir.Page, saved string, s *syntax.Str) {
	target := findPage(part, s.Value)
	switch {
	case target == nil:
		c.d.Errorf(s.Pos, "there is no page at %s", s.Value)
		c.d.HintNames(s.Value, pageRoutes(part), "pages")
	case !strings.Contains(s.Value, "[id]"):
		pg.Then = s.Value
	case saved == "":
		c.d.Errorf(s.Pos, "%s returns nothing, so there's no [id] to go to", pg.Form)
	case !target.ByID || target.Entity != saved:
		c.d.Errorf(s.Pos, "[id] here is the saved %s's, but %s doesn't load a %s", saved, s.Value, saved)
	default:
		pg.Then = s.Value
	}
}

func findResource(backend *ir.Part, name string) *ir.Resource {
	i := slices.IndexFunc(backend.Resources, func(r *ir.Resource) bool { return r.Name == name })
	if i < 0 {
		return nil
	}
	return backend.Resources[i]
}

func resourceNames(backend *ir.Part) []string {
	var names []string
	for _, r := range backend.Resources {
		names = append(names, r.Name)
	}
	return names
}

func findPage(part *ir.Part, route string) *ir.Page {
	i := slices.IndexFunc(part.Pages, func(p *ir.Page) bool { return p.Route == route })
	if i < 0 {
		return nil
	}
	return part.Pages[i]
}

func pageRoutes(part *ir.Part) []string {
	var routes []string
	for _, p := range part.Pages {
		routes = append(routes, p.Route)
	}
	return routes
}
