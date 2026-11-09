package check

import (
	"slices"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

var primitives = map[string]ir.Kind{
	"string": ir.String, "slug": ir.Slug, "email": ir.Email, "markdown": ir.Markdown,
	"int": ir.Int, "bool": ir.Bool, "date": ir.Date, "datetime": ir.DateTime,
}

var primitiveNames = []string{"string", "slug", "email", "markdown", "int", "bool", "date", "datetime"}

func (c *checker) declareEntity(e *syntax.Entity) {
	if !c.declareType(e.Name) {
		return
	}
	ent := &ir.Entity{Name: e.Name.Name}
	c.entities[ent.Name] = ent
	c.members[ent.Name] = map[string]member{"id": {e.Name.Pos, "the automatic id"}}
	c.app.Entities = append(c.app.Entities, ent)
}

func (c *checker) fields(ent *ir.Entity, e *syntax.Entity) {
	if len(e.Fields) == 0 {
		c.d.Errorf(e.Name.Pos, "entity %s has no fields", ent.Name)
	}
	for _, f := range e.Fields {
		if !c.claim(ent.Name, f.Name, "field "+f.Name.Name) {
			continue
		}
		if fd := c.field(ent, e, f); fd != nil {
			ent.Fields = append(ent.Fields, fd)
		}
	}
}

func (c *checker) field(ent *ir.Entity, e *syntax.Entity, f *syntax.Field) *ir.Field {
	fd := &ir.Field{Name: f.Name.Name}
	kind, primitive := primitives[f.Type.Name]
	switch {
	case primitive:
		fd.Kind = kind
	case c.app.Enum(f.Type.Name) != nil:
		fd.Kind, fd.Enum = ir.EnumValue, f.Type.Name
	case c.entities[f.Type.Name] != nil:
		c.d.Errorf(f.Type.Pos, "a field can't hold an entity; use a relation")
		c.d.Hint("`%s belongs to %s: %s`", ent.Name, f.Name.Name, f.Type.Name)
		return nil
	default:
		c.d.Errorf(f.Type.Pos, "unknown type '%s'", f.Type.Name)
		c.d.HintNames(f.Type.Name, append(slices.Clone(primitiveNames), enumNames(c.app)...), "types")
		return nil
	}
	c.limits(fd, f)
	c.slugFrom(fd, e, f)
	c.modifiers(fd, f)
	return fd
}

func enumNames(a *ir.App) []string {
	var out []string
	for _, e := range a.Enums {
		out = append(out, e.Name)
	}
	return out
}

func (c *checker) limits(fd *ir.Field, f *syntax.Field) {
	for _, l := range f.Limits {
		if fd.Kind != ir.String && fd.Kind != ir.Int {
			c.d.Errorf(l.Name.Pos, "(min/max) only applies to string and int fields")
			return
		}
		v := l.Value
		switch l.Name.Name {
		case "min":
			fd.Min = &v
		case "max":
			fd.Max = &v
		default:
			c.d.Errorf(l.Name.Pos, "unknown limit '%s'; limits are min and max", l.Name.Name)
		}
	}
	if fd.Min != nil && fd.Max != nil && *fd.Min > *fd.Max {
		c.d.Errorf(f.Limits[0].Name.Pos, "min is larger than max")
	}
}

func (c *checker) slugFrom(fd *ir.Field, e *syntax.Entity, f *syntax.Field) {
	if fd.Kind == ir.Slug {
		fd.Unique = true
	}
	if f.From == nil {
		return
	}
	if fd.Kind != ir.Slug {
		c.d.Errorf(f.From.Pos, "only slug fields can be made `from` another field")
		return
	}
	i := slices.IndexFunc(e.Fields, func(x *syntax.Field) bool { return x.Name.Name == f.From.Name })
	switch {
	case i < 0:
		c.d.Errorf(f.From.Pos, "%s has no field '%s'", e.Name.Name, f.From.Name)
	case e.Fields[i].Type.Name != "string":
		c.d.Errorf(f.From.Pos, "a slug can only be made from a string field")
	default:
		fd.SlugFrom = f.From.Name
	}
}

func (c *checker) modifiers(fd *ir.Field, f *syntax.Field) {
	for _, m := range f.Mods {
		switch m.Name {
		case "unique":
			if fd.Kind == ir.Markdown || fd.Kind == ir.Bool {
				c.d.Errorf(m.Pos, "%s fields can't be unique", f.Type.Name)
			}
			fd.Unique = true
		case "optional":
			fd.Optional = true
		default:
			c.d.Errorf(m.Pos, "unknown modifier '%s'", m.Name)
			c.d.HintNames(m.Name, []string{"unique", "optional"}, "modifiers")
		}
	}
}

// claim reserves a name on an entity, reporting a clash with what already has it.
func (c *checker) claim(entity string, n syntax.Ident, what string) bool {
	return c.claimAt(entity, n.Name, n, what)
}

func (c *checker) claimAt(entity, name string, at syntax.Ident, what string) bool {
	m := c.members[entity]
	if prev, ok := m[name]; ok {
		c.d.Errorf(at.Pos, "%s.%s is already defined by %s", entity, name, prev.what)
		c.d.Hint("see %s", prev.pos)
		return false
	}
	m[name] = member{at.Pos, what}
	return true
}
