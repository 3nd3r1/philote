package check

import (
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// maxCopy caps UI copy like empty and not_found (D-006).
const maxCopy = 200

// views checks a frontend's components, then the pages that use them, then
// the links between them.
func (c *checker) views(f *frontendSrc, backend *ir.Part) {
	links := map[*ir.Component]*syntax.Str{}
	for _, s := range f.components {
		if comp := c.component(f.part, backend, s); comp != nil && s.LinksTo != nil {
			links[comp] = s.LinksTo
		}
	}
	for _, s := range f.pages {
		c.page(f.part, backend, s)
	}
	for _, comp := range f.part.Components {
		if s, ok := links[comp]; ok {
			c.link(f.part, comp, s)
		}
	}
}

func (c *checker) component(part, backend *ir.Part, s *syntax.Component) *ir.Component {
	name := s.Name.Name
	switch {
	case name[0] < 'A' || name[0] > 'Z':
		c.d.Errorf(s.Name.Pos, "component names start with a capital letter, like HorseCard")
		return nil
	case findComponent(part, name) != nil:
		c.d.Errorf(s.Name.Pos, "there is already a component named %s", name)
		return nil
	case len(s.Params) != 1:
		c.d.Errorf(s.Name.Pos, "a component takes one record, like (horse: Horse)")
		return nil
	}
	ent := c.lookupEntity(s.Params[0].Type)
	if ent == nil {
		return nil
	}
	comp := &ir.Component{Name: name, Param: s.Params[0].Name.Name, Entity: ent.Name}
	if len(s.Shows) == 0 {
		c.d.Errorf(s.Name.Pos, "%s doesn't show anything", name)
		c.d.Hint("add `shows: %s.<field>`", comp.Param)
	}
	for _, sh := range s.Shows {
		show := c.show(backend, comp, ent, sh)
		switch {
		case show == nil:
		case slices.ContainsFunc(comp.Shows, func(x *ir.Show) bool { return x.Name == show.Name }):
			c.d.Errorf(sh.Path[1].Pos, "%s.%s is shown twice", comp.Param, show.Name)
		default:
			comp.Shows = append(comp.Shows, show)
		}
	}
	part.Components = append(part.Components, comp)
	return comp
}

// show resolves one path like horse.name or match.first.
func (c *checker) show(backend *ir.Part, comp *ir.Component, ent *ir.Entity, s syntax.Show) *ir.Show {
	head := s.Path[0]
	switch {
	case head.Name != comp.Param:
		c.d.Errorf(head.Pos, "'%s' isn't %s's parameter; start with %s", head.Name, comp.Name, comp.Param)
		return nil
	case len(s.Path) != 2:
		c.d.Errorf(head.Pos, "show one level below %s, like %s.<field>", comp.Param, comp.Param)
		c.d.Hint("a relation already shows its target's display field")
		return nil
	}
	m := s.Path[1]
	if f := findField(ent, m.Name); f != nil {
		if f.Kind == ir.Markdown {
			c.d.Errorf(m.Pos, "markdown fields can't be shown yet")
			return nil
		}
		return c.format(&ir.Show{Name: f.Name, Field: f}, s)
	}
	rel := findRelation(ent, m.Name)
	if rel == nil {
		c.d.Errorf(m.Pos, "%s has no field or relation '%s'", ent.Name, m.Name)
		c.d.HintNames(m.Name, memberNames(ent), "fields and relations")
		return nil
	}
	show := c.format(&ir.Show{Name: rel.Name, Relation: rel}, s)
	if show == nil {
		return nil
	}
	if show.Display = c.entities[rel.Target].Display; show.Display == "" {
		c.d.Errorf(m.Pos, "%s has no display field, so %s.%s can't be shown", rel.Target, comp.Param, rel.Name)
		c.d.Hint("add `display: <field>` to %s, like `display: name`", rel.Target)
		return nil
	}
	// A belongs-to read gets the target by id; a has-many read lists it
	// from the parent's relation route.
	loads := rel.Target
	if rel.Many {
		loads = ent.Name
	}
	if show.Resource = c.loader(backend, loads, m); show.Resource == "" {
		return nil
	}
	return show
}

// format checks `as <format>`: chips is the only one, and only for has-many
// relations, which need it.
func (c *checker) format(show *ir.Show, s syntax.Show) *ir.Show {
	many := show.Relation != nil && show.Relation.Many
	switch {
	case s.Format != nil && s.Format.Name != "chips":
		c.d.Errorf(s.Format.Pos, "unknown format '%s'", s.Format.Name)
		c.d.HintNames(s.Format.Name, []string{"chips"}, "formats")
	case s.Format != nil && !many:
		c.d.Errorf(s.Format.Pos, "only a has-many relation can be shown as chips")
	case s.Format == nil && many:
		c.d.Errorf(s.Path[1].Pos, "%s.%s is a list; show it `as chips`", s.Path[0].Name, s.Path[1].Name)
	default:
		return show
	}
	return nil
}

// loader finds the one resource in the backend that serves entity with get.
func (c *checker) loader(backend *ir.Part, entity string, at syntax.Ident) string {
	var names []string
	for _, r := range backend.Resources {
		if r.Entity == entity && slices.Contains(r.Ops, ir.Get) {
			names = append(names, r.Name)
		}
	}
	switch len(names) {
	case 0:
		c.d.Errorf(at.Pos, "%s has no resource that gets %s, so this can't be loaded", backend.Name, entity)
		c.d.Hint("add `resource <name>: %s` to %s", entity, backend.Name)
		return ""
	case 1:
		return names[0]
	}
	c.d.Errorf(at.Pos, "several resources in %s get %s: %s", backend.Name, entity, strings.Join(names, ", "))
	return ""
}

func findComponent(part *ir.Part, name string) *ir.Component {
	i := slices.IndexFunc(part.Components, func(x *ir.Component) bool { return x.Name == name })
	if i < 0 {
		return nil
	}
	return part.Components[i]
}

func findField(ent *ir.Entity, name string) *ir.Field {
	i := slices.IndexFunc(ent.Fields, func(f *ir.Field) bool { return f.Name == name })
	if i < 0 {
		return nil
	}
	return ent.Fields[i]
}

func findRelation(ent *ir.Entity, name string) *ir.Relation {
	i := slices.IndexFunc(ent.Relations, func(r *ir.Relation) bool { return r.Name == name })
	if i < 0 {
		return nil
	}
	return ent.Relations[i]
}

func memberNames(ent *ir.Entity) []string {
	out := fieldNames(ent)
	for _, r := range ent.Relations {
		out = append(out, r.Name)
	}
	return out
}
