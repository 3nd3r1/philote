package check

import (
	"slices"
	"strconv"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// actions resolves a component's actions once every page is known (D-020).
func (c *checker) actions(part, backend *ir.Part, comp *ir.Component, src []syntax.Action) {
	for _, s := range src {
		a := &ir.Action{}
		ok := c.actionKind(part, backend, comp, a, s)
		if ok && s.Then != nil {
			a.Then, a.ThenBy, ok = c.recordRoute(part, comp, s.Then)
		}
		if ok && s.Label != nil {
			if len(s.Label.Value) > maxCopy {
				c.d.Errorf(s.Label.Pos, "UI text is capped at %d characters", maxCopy)
				continue
			}
			a.Label = s.Label.Value
		}
		if ok {
			comp.Actions = append(comp.Actions, a)
		}
	}
}

func (c *checker) actionKind(part, backend *ir.Part, comp *ir.Component, a *ir.Action, s syntax.Action) bool {
	switch s.Verb.Name {
	case "open":
		if s.Then != nil {
			c.d.Errorf(s.Then.Pos, "open already goes to a page; then doesn't apply")
			return false
		}
		var ok bool
		a.Kind = ir.Open
		a.Route, a.RouteBy, ok = c.recordRoute(part, comp, s.Route)
		return ok
	case "delete":
		if s.Binds != nil {
			c.d.Errorf(s.Verb.Pos, "delete takes no inputs; it deletes the %s it's on", comp.Param)
			return false
		}
		res := c.serving(backend, comp.Entity, ir.Delete, "deletes", s.Verb)
		a.Kind, a.Endpoint = ir.Remove, res+".delete"
		return res != ""
	}
	i := slices.IndexFunc(backend.Custom, func(e *ir.Endpoint) bool { return e.Name == s.Verb.Name })
	if i < 0 {
		var names []string
		for _, e := range backend.Custom {
			names = append(names, e.Name)
		}
		c.d.Errorf(s.Verb.Pos, "unknown action '%s'", s.Verb.Name)
		c.d.HintNames(s.Verb.Name, append([]string{"open", "delete"}, names...), "actions")
		return false
	}
	ep := backend.Custom[i]
	a.Kind, a.Endpoint = ir.Call, ep.Name
	return c.binds(comp, ep, a, s)
}

// binds checks that an action gives every input of its endpoint a value of
// the right type, and nothing else. Inputs can't come from who's signed in
// yet (Q-004).
func (c *checker) binds(comp *ir.Component, ep *ir.Endpoint, a *ir.Action, s syntax.Action) bool {
	given := map[string]syntax.Bind{}
	for _, b := range s.Binds {
		switch {
		case !slices.ContainsFunc(ep.Flow.Input, func(v *ir.Value) bool { return v.Name == b.Name.Name }):
			c.d.Errorf(b.Name.Pos, "%s takes no input named '%s'", ep.Name, b.Name.Name)
			return false
		case given[b.Name.Name].Name.Name != "":
			c.d.Errorf(b.Name.Pos, "'%s' is given twice", b.Name.Name)
			return false
		}
		given[b.Name.Name] = b
	}
	ok := true
	for _, v := range ep.Flow.Input {
		b, set := given[v.Name]
		if !set {
			c.d.Errorf(s.Verb.Pos, "%s needs '%s'", ep.Name, v.Name)
			c.d.Hint("give it from the %s, like %s(%s: %s), or make %s a form page", comp.Param, ep.Name, v.Name, comp.Param, ep.Name)
			ok = false
			continue
		}
		bind := c.bind1(comp, v, b)
		if bind == nil {
			ok = false
			continue
		}
		a.Binds = append(a.Binds, bind)
	}
	return ok
}

// bind1 checks one input's value: the record, one of its fields or
// belongs-to relations, or a literal.
func (c *checker) bind1(comp *ir.Component, v *ir.Value, b syntax.Bind) *ir.Bind {
	out := &ir.Bind{Input: v.Name}
	switch {
	case b.Int != nil:
		n, _ := strconv.Atoi(b.Int.Name)
		if v.Field == nil || v.Field.Kind != ir.Int || !inBounds(v.Field, n) {
			c.d.Errorf(b.Int.Pos, "'%s' is %s, not this number", v.Name, typeName(v))
			return nil
		}
		out.Literal = b.Int.Name
	case b.Str != nil:
		if v.Field == nil || !stringKinds[v.Field.Kind] || (v.Field.Max != nil && len(b.Str.Value) > *v.Field.Max) {
			c.d.Errorf(b.Str.Pos, "'%s' is %s, not this text", v.Name, typeName(v))
			return nil
		}
		out.Literal = strconv.Quote(b.Str.Value)
	case len(b.Path) == 1 && b.Path[0].Name != comp.Param:
		lit := b.Path[0].Name
		if !c.literalWord(v, lit) {
			c.d.Errorf(b.Path[0].Pos, "'%s' is %s; '%s' isn't the %s or one of its values", v.Name, typeName(v), lit, comp.Param)
			return nil
		}
		out.Literal = lit
		if v.Field.Kind == ir.EnumValue {
			out.Literal = strconv.Quote(lit)
		}
	default:
		out.Field = c.bindPath(comp, v, b.Path)
		if out.Field == "" {
			return nil
		}
	}
	return out
}

// stringKinds are the field kinds a string literal can be.
var stringKinds = map[ir.Kind]bool{ir.String: true, ir.Slug: true, ir.Email: true, ir.Markdown: true, ir.Date: true, ir.DateTime: true}

// literalWord accepts true and false for a bool, and an enum's values.
func (c *checker) literalWord(v *ir.Value, word string) bool {
	switch {
	case v.Field == nil:
		return false
	case v.Field.Kind == ir.Bool:
		return word == "true" || word == "false"
	case v.Field.Kind == ir.EnumValue:
		return slices.Contains(c.app.Enum(v.Field.Enum).Values, word)
	}
	return false
}

func inBounds(f *ir.Field, n int) bool {
	return (f.Min == nil || n >= *f.Min) && (f.Max == nil || n <= *f.Max)
}

// bindPath resolves book, book.title or book.author to the field of the
// record the action sends.
func (c *checker) bindPath(comp *ir.Component, v *ir.Value, path []syntax.Ident) string {
	head := path[0]
	switch {
	case head.Name != comp.Param:
		c.d.Errorf(head.Pos, "'%s' isn't %s's parameter; start with %s", head.Name, comp.Name, comp.Param)
		return ""
	case len(path) > 2:
		c.d.Errorf(head.Pos, "use %s or one level below it", comp.Param)
		return ""
	case len(path) == 1 && v.Entity != comp.Entity:
		c.d.Errorf(head.Pos, "'%s' is %s, but %s is a %s", v.Name, typeName(v), comp.Param, comp.Entity)
		return ""
	case len(path) == 1:
		return "id"
	}
	ent, m := c.entities[comp.Entity], path[1]
	if f := findField(ent, m.Name); f != nil {
		if v.Field == nil || v.Field.Kind != f.Kind || v.Field.Enum != f.Enum {
			c.d.Errorf(m.Pos, "'%s' is %s, but %s.%s is %s", v.Name, typeName(v), comp.Param, f.Name, typeName(&ir.Value{Field: f}))
			return ""
		}
		return f.Name
	}
	rel := findRelation(ent, m.Name)
	switch {
	case rel == nil:
		c.d.Errorf(m.Pos, "%s has no field or relation '%s'", ent.Name, m.Name)
		c.d.HintNames(m.Name, memberNames(ent), "fields and relations")
	case rel.Many:
		c.d.Errorf(m.Pos, "%s.%s is a list; an input takes one record", comp.Param, rel.Name)
	case v.Entity != rel.Target:
		c.d.Errorf(m.Pos, "'%s' is %s, but %s.%s is a %s", v.Name, typeName(v), comp.Param, rel.Name, rel.Target)
	default:
		return rel.FK
	}
	return ""
}
