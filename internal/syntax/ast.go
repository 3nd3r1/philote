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

// Page is `page "<route>" { data: <resource> [by id]  layout: [list of] <Component>
// empty: "..."  not_found: "..." }`, or a form: `page "<route>" { form:
// <resource>.<op> [by id]  then: "<route>"  not_found: "..." }`.
type Page struct {
	Word     Ident
	Route    *Str
	Data     *Ident
	By       *Ident // the field after `by`
	List     bool   // layout: list of ...
	Layout   *Ident
	Empty    *Str
	NotFound *Str
	Form     *Ident // a resource, or a custom endpoint when FormOp is nil
	FormOp   *Ident // the operation after the resource's '.'
	Then     *Str
}

// Component is `component <Name>(<param>: <Entity>) { shows: <path> [as <format>], ...
// links_to: "<route>"  look: "..." }`. Its body is written by a model.
type Component struct {
	Name    Ident
	Params  []*Field
	Shows   []Show
	LinksTo *Str
	Look    *Str
	Actions []Action
}

// Action is something a component's record can have done to it: `open
// "<route>"` or `delete`, then an optional `then "<route>"` and `as "<label>"`.
type Action struct {
	Verb  Ident // open or delete
	Route *Str  // for open
	Then  *Str
	Label *Str
}

// Show is one thing a component shows: a path like horse.name, with an
// optional format like `as chips`.
type Show struct {
	Path   []Ident
	Format *Ident
}

func (a *App) Start() diag.Pos       { return a.Name.Pos }
func (e *Endpoint) Start() diag.Pos  { return e.Name.Pos }
func (s *Step) Start() diag.Pos      { return s.Name.Pos }
func (p *Part) Start() diag.Pos      { return p.Kind.Pos }
func (e *Enum) Start() diag.Pos      { return e.Name.Pos }
func (e *Entity) Start() diag.Pos    { return e.Name.Pos }
func (r *Relation) Start() diag.Pos  { return r.From.Pos }
func (r *Resource) Start() diag.Pos  { return r.Name.Pos }
func (u *Uses) Start() diag.Pos      { return u.Word.Pos }
func (p *Page) Start() diag.Pos      { return p.Word.Pos }
func (c *Component) Start() diag.Pos { return c.Name.Pos }
