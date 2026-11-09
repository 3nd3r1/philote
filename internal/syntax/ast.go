package syntax

import "github.com/3nd3r1/philote/internal/diag"

// Ident is a name and where it was written.
type Ident struct {
	Name string
	Pos  diag.Pos
}

// File is one parsed source file.
type File struct {
	Name  string
	Items []Item
}

// Item is anything that can appear as a statement: a block or a relation.
type Item interface{ Start() diag.Pos }

// App is `app <name> { ... }`.
type App struct {
	Name  Ident
	Items []Item
}

// Part is `backend <name>: <target>` or `frontend <name>: <target>`,
// with a body inline or loaded from <name>.philote.
type Part struct {
	Kind    Ident
	Name    Ident
	Target  Ident
	HasBody bool
	Body    []Item
}

// Enum is `enum <Name> { a, b, c }`.
type Enum struct {
	Name   Ident
	Values []Ident
}

// Entity is `entity <Name> { field: type ... }`.
type Entity struct {
	Name   Ident
	Fields []*Field
}

// Field is `<name>: <type>(limits) [from <field>] [modifiers]`.
type Field struct {
	Name   Ident
	Type   Ident
	Limits []Limit
	From   *Ident
	Mods   []Ident
}

// Limit is one of `(min N, max N)`.
type Limit struct {
	Name  Ident
	Value int
}

// Relation is `<From> has many <name>: <To>` or `<From> belongs to <name>: <To>`,
// with an optional `(back <name>)` naming the reverse side.
type Relation struct {
	From Ident
	Verb Ident // "has many" or "belongs to", positioned at the first word
	Name Ident
	To   Ident
	Back *Ident
}

// Resource is `resource <name>: <Entity> [{ only: list, get }]`.
type Resource struct {
	Name   Ident
	Entity Ident
	Only   []Ident
}

func (a *App) Start() diag.Pos      { return a.Name.Pos }
func (p *Part) Start() diag.Pos     { return p.Kind.Pos }
func (e *Enum) Start() diag.Pos     { return e.Name.Pos }
func (e *Entity) Start() diag.Pos   { return e.Name.Pos }
func (r *Relation) Start() diag.Pos { return r.From.Pos }
func (r *Resource) Start() diag.Pos { return r.Name.Pos }
