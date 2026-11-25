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
	List   bool // `list of`, for a component parameter or a step's parameter or result
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
	Name    Ident
	Method  *Ident
	Path    *Str
	Input   []*Field
	Output  *Ident
	List    bool     // output: list of <Entity>
	Outputs []*Field // output: { <name>: <Type>, ... }, the values it returns by name
	Errors  []Ident
	Steps   []Ident
}

// Step is `step <name>(<params>) [-> <name>: <Type>] { fails: a, b  reads: A
// writes: B  intent: "..." }`. Its body is written by a model.
type Step struct {
	Name   Ident
	Params []*Field
	Result *Field
	Fails  []Ident
	Reads  []Ident
	Writes []Ident
	Intent *Str
}

// Uses is `uses: <part>`, naming the backend a frontend calls.
type Uses struct {
	Word Ident
	Part Ident
}

// Look is `look: "<text>"` in a frontend: how the whole app should look,
// for the model that writes its theme.
type Look struct {
	Word  Ident
	Value *Str
}

// Theme is `theme: <name>` in a frontend: one of the fixed built-in themes.
type Theme struct {
	Word Ident
	Name Ident
}

// Viewer is `viewer: <Entity>` in a frontend: the record a visitor picks
// to act as, kept in a cookie until auth replaces the picker.
type Viewer struct {
	Word   Ident
	Entity Ident
}

// Nav is `nav: <place>, "<route>" [as "<label>"], ...` in a frontend:
// where the app's nav sits, and the pages it links to, in order.
type Nav struct {
	Word  Ident
	Place Ident
	Items []NavItem
}

// NavItem is one page in the nav, with its text when the page's title
// isn't it.
type NavItem struct {
	Route *Str
	Label *Str
}

// Page is `page "<route>" { load: <name> [from <resource>] [by id]
// shows: <section>, ...  empty: "..."  not_found: "..." }`, or a form:
// `page "<route>" { form: <resource>.<op> [by id]  then: "<route>"
// not_found: "..." }`.
type Page struct {
	Word     Ident
	Route    *Str
	Load     *PageLoad
	Shows    []Section
	Empty    *Str
	NotFound *Str
	Form     *Ident // a resource, or a custom endpoint when FormOp is nil
	FormOp   *Ident // the operation after the resource's '.'
	By       *Ident // a form's `by <field>`
	Then     *Str
}

// PageLoad is `<name> [from <resource>] [by <field>]`, or `<name> from
// <endpoint>(<input>: <value>, ...)`: what a page fetches. Without from, the
// resource is the name.
type PageLoad struct {
	Name Ident
	From *Ident
	By   *Ident
	Args []Arg // set for an endpoint, even when its parentheses are empty
	Call bool  // from names an endpoint called with Args
}

// Arg is `<input>: <value>` or a bare `<input>`, binding one of an
// endpoint's inputs in a load or a component's calls.
type Arg struct {
	Name  Ident
	Value *Ident // nil when the input is left open
}

// Call is `<endpoint>(<arg>, ...)`: an endpoint a component calls.
type Call struct {
	Endpoint Ident
	Args     []Arg
}

// Section is one thing a page shows: `[list of] <Component>[(<load>)]`.
type Section struct {
	List      bool
	Component Ident
	Arg       *Ident
}

// Component is `component <Name>[(<param>: [list of] <Entity>)] { shows:
// <path> [as <format>], ...  says: "...", ...  links_to: "<route>" [as
// "<label>"]  calls: <endpoint>(<arg>, ...)  look: "..." }`. Its body is
// written by a model.
type Component struct {
	Name      Ident
	Params    []*Field
	Kind      *Ident
	Shows     []Show
	Says      []*Str
	LinksTo   *Str
	LinkLabel *Str
	Look      *Str
	Actions   []Action
	Picks     *Ident // `picks: viewer then "<route>"`: tapping a record makes it the viewer
	PicksThen *Str
	Calls     *Call
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
func (l *Look) Start() diag.Pos      { return l.Word.Pos }
func (t *Theme) Start() diag.Pos     { return t.Word.Pos }
func (n *Nav) Start() diag.Pos       { return n.Word.Pos }
func (v *Viewer) Start() diag.Pos    { return v.Word.Pos }
func (p *Page) Start() diag.Pos      { return p.Word.Pos }
func (c *Component) Start() diag.Pos { return c.Name.Pos }
