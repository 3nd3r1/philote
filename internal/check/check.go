package check

import (
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/diag"
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// Targets maps each available target name to the kind of part it builds.
type Targets map[string]ir.PartKind

// member is a name on an entity (field, relation or foreign key) and what declared it.
type member struct {
	pos  diag.Pos
	what string
}

type checker struct {
	d        *diag.List
	targets  Targets
	app      *ir.App
	types    map[string]diag.Pos
	entities map[string]*ir.Entity
	members  map[string]map[string]member
	uses     map[*ir.Part]syntax.Ident // frontend -> the backend its `uses` names
}

// Run checks a loaded source and returns its IR. The IR is only meaningful
// when d has no errors afterwards.
func Run(src *syntax.Source, targets Targets, d *diag.List) *ir.App {
	c := &checker{
		d:        d,
		targets:  targets,
		types:    map[string]diag.Pos{},
		entities: map[string]*ir.Entity{},
		members:  map[string]map[string]member{},
		uses:     map[*ir.Part]syntax.Ident{},
	}
	if src.Root == nil {
		return nil
	}
	app, enums, entities, relations := c.topLevel(src.Root)
	if app == nil {
		return nil
	}
	c.app = &ir.App{Name: app.Name.Name}
	for _, e := range enums {
		c.enum(e)
	}
	for _, e := range entities {
		c.declareEntity(e)
	}
	for _, e := range entities {
		if ent := c.entities[e.Name.Name]; ent != nil {
			c.fields(ent, e)
		}
	}
	for _, r := range relations {
		c.relation(r)
	}
	c.parts(app)
	return c.app
}

func (c *checker) topLevel(f *syntax.File) (app *syntax.App, enums []*syntax.Enum, entities []*syntax.Entity, relations []*syntax.Relation) {
	for _, it := range f.Items {
		switch x := it.(type) {
		case *syntax.App:
			if app != nil {
				c.d.Errorf(x.Name.Pos, "only one app is allowed")
				continue
			}
			app = x
		case *syntax.Enum:
			enums = append(enums, x)
		case *syntax.Entity:
			entities = append(entities, x)
		case *syntax.Relation:
			relations = append(relations, x)
		default:
			c.d.Errorf(it.Start(), "this belongs inside a part, not at the top level")
		}
	}
	if app == nil {
		c.d.Errorf(diag.Pos{File: f.Name, Line: 1, Col: 1}, "missing app block")
		c.d.Hint("start with `app <name> { backend api: hono }`")
	}
	return app, enums, entities, relations
}

// declareType claims a type name for an enum or entity.
func (c *checker) declareType(n syntax.Ident) bool {
	if prev, ok := c.types[n.Name]; ok {
		c.d.Errorf(n.Pos, "'%s' is already declared", n.Name)
		c.d.Hint("first declared at %s", prev)
		return false
	}
	if _, ok := primitives[n.Name]; ok {
		c.d.Errorf(n.Pos, "'%s' is a built-in type", n.Name)
		return false
	}
	c.types[n.Name] = n.Pos
	return true
}

func (c *checker) enum(e *syntax.Enum) {
	if !c.declareType(e.Name) {
		return
	}
	en := &ir.Enum{Name: e.Name.Name}
	for _, v := range e.Values {
		if slices.Contains(en.Values, v.Name) {
			c.d.Errorf(v.Pos, "'%s' is listed twice", v.Name)
			continue
		}
		en.Values = append(en.Values, v.Name)
	}
	c.app.Enums = append(c.app.Enums, en)
}

func (c *checker) entityNames() []string {
	var out []string
	for _, e := range c.app.Entities {
		out = append(out, e.Name)
	}
	return out
}

func (c *checker) lookupEntity(n syntax.Ident) *ir.Entity {
	e := c.entities[n.Name]
	if e == nil {
		c.d.Errorf(n.Pos, "unknown entity '%s'", n.Name)
		c.d.HintNames(n.Name, c.entityNames(), "entities")
	}
	return e
}

// snake turns OrderLine into order_line.
func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}
