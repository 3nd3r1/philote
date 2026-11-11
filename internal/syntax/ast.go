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

// Entity is `entity <Name> { [display: <field>] field: type ... }`.
type Entity struct {
	Name     Ident
	Displays []Display // more than one is a check error
	Fields   []*Field
}

// Display is `display: <field>`, the field that stands for a record when
// another record shows it.
type Display struct {
	Word  Ident
	Field Ident
}

// Field is `<name>: <type>(limits) [from <field>] [modifiers]`. Inputs and
// step parameters use the same shape without from or modifiers.
type Field struct {
	Name   Ident
	Type   Ident
	Limits []Limit
	From   *Ident
	Mods   []Ident
}

// Str is a quoted string and where it was written.
type Str struct {
	Value string
	Pos   diag.Pos
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

// Endpoint is a custom route whose behaviour is a list of steps:
// `endpoint <name> { route: POST "/path"  input: { ... }  output: <Entity>  errors: a, b  steps: x, y }`.
type Endpoint struct {
	Name   Ident
	Method *Ident
	Path   *Str
	Input  []*Field
	Output *Ident
	Errors []Ident
	Steps  []Ident
}

// Step is `step <name>(<params>) [-> <name>: <Type>] { fails: a, b  intent: "..." }`.
// Its body is written by a model.
type Step struct {
	Name   Ident
	Params []*Field
	Result *Field
	Fails  []Ident
	Intent *Str
}

// Uses is `uses: <part>`, naming the backend a frontend calls.
type Uses struct {
	Word Ident
	Part Ident
}

func (a *App) Start() diag.Pos      { return a.Name.Pos }
func (e *Endpoint) Start() diag.Pos { return e.Name.Pos }
func (s *Step) Start() diag.Pos     { return s.Name.Pos }
func (p *Part) Start() diag.Pos     { return p.Kind.Pos }
func (e *Enum) Start() diag.Pos     { return e.Name.Pos }
func (e *Entity) Start() diag.Pos   { return e.Name.Pos }
func (r *Relation) Start() diag.Pos { return r.From.Pos }
func (r *Resource) Start() diag.Pos { return r.Name.Pos }
func (u *Uses) Start() diag.Pos     { return u.Word.Pos }
