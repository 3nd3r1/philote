package spec

// Typed AST. Every name carries its position so errors point at the source.

type Ident struct {
	Name string
	Pos  Pos
}

type Str struct {
	Pos   Pos
	Value string
}

type Int struct {
	Pos   Pos
	Value int
}

type File struct {
	Path       string
	Src        string
	Decls      []any // *App, *Entity, *Page, *Component in source order
	App        *App
	Entities   []*Entity
	Pages      []*Page
	Components []*Component
}

type App struct {
	Pos    Pos
	Name   Ident
	Target *Ident
	Style  *Str // e.g. "tokens/minimal"
}

type Entity struct {
	Pos    Pos
	Name   Ident
	Fields []*Field
}

func (e *Entity) Field(name string) *Field {
	for _, f := range e.Fields {
		if f.Name.Name == name {
			return f
		}
	}
	return nil
}

func (e *Entity) FieldNames() []string {
	var out []string
	for _, f := range e.Fields {
		out = append(out, f.Name.Name)
	}
	return out
}

type Field struct {
	Name     Ident
	Type     TypeRef
	SlugFrom *Ident
	Max      *Int
	Unique   bool
	Optional bool
}

type TypeRef struct {
	Pos  Pos
	Name string
	List bool
}

type Page struct {
	Pos      Pos
	Route    Route
	Data     *Query
	Layout   *Layout
	Empty    *Str
	NotFound *Int
}

type Route struct {
	Pos  Pos
	Raw  string
	Segs []Seg
}

// Seg is a route segment: a literal ("blog") or a parameter ("[slug]").
type Seg struct {
	Pos   Pos
	Lit   string
	Param string
}

type Query struct {
	Pos    Pos
	Entity Ident
	By     *Ident
	Where  []*Cond
	Order  *Order
	Limit  *Int
}

type Cond struct {
	Field Ident
	Op    string
	OpPos Pos
	Value Expr
}

type Order struct {
	Field Ident
	Desc  bool
}

type Layout struct {
	Pos       Pos
	List      bool
	Component Ident
}

type Component struct {
	Pos     Pos
	Name    Ident
	Params  []*Param
	Native  bool
	From    *Str
	Shows   []*Show
	LinksTo *Template
}

type Param struct {
	Name Ident
	Type TypeRef
}

type Show struct {
	Expr Expr
	As   *Ident
}

// Template is a link like "/blog/[post.slug]".
type Template struct {
	Pos  Pos
	Raw  string
	Segs []TSeg
}

type TSeg struct {
	Pos  Pos
	Lit  string
	Path *PathExpr
}

type Expr interface{ exprPos() Pos }

type PathExpr struct{ Parts []Ident }
type CallExpr struct {
	Fn   Ident
	Args []Expr
}
type IntLit struct {
	Pos   Pos
	Value int
}
type StrLit struct {
	Pos   Pos
	Value string
}

func (e *PathExpr) exprPos() Pos { return e.Parts[0].Pos }
func (e *CallExpr) exprPos() Pos { return e.Fn.Pos }
func (e *IntLit) exprPos() Pos   { return e.Pos }
func (e *StrLit) exprPos() Pos   { return e.Pos }
