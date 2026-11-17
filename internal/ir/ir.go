package ir

import "github.com/3nd3r1/philote/internal/diag"

// App is the root of the checked tree.
type App struct {
	Name     string
	Enums    []*Enum
	Entities []*Entity
	Parts    []*Part
}

type Enum struct {
	Name   string
	Values []string
}

// Kind is a field's scalar type.
type Kind int

const (
	String Kind = iota
	Slug
	Email
	Markdown
	Int
	Bool
	Date
	DateTime
	EnumValue
)

type Entity struct {
	Name      string
	Display   string // the field shown when another record shows this one
	Fields    []*Field
	Relations []*Relation
}

type Field struct {
	Name     string
	Kind     Kind
	Enum     string // for EnumValue
	Min, Max *int
	SlugFrom string // for Slug fields made from another field
	Unique   bool
	Optional bool
	Readonly bool // only steps write it, so it's left out of resource input (D-023)
}

// Relation is one side of a relation, as seen from the entity that has it.
type Relation struct {
	Name   string
	Many   bool   // has many; otherwise this side holds the foreign key
	Target string // entity name
	FK     string // foreign key field on this entity, for the one side
	Key    string // foreign key field on Target that points back, for the many side
}

// PartKind says which side of the app a part is.
type PartKind int

const (
	Backend PartKind = iota
	Frontend
)

// Part is a backend or frontend node, built by one target.
type Part struct {
	Name       string
	Kind       PartKind
	Target     string
	Uses       string // for a frontend: the backend part it calls
	Resources  []*Resource
	Custom     []*Endpoint // endpoints declared with steps
	Steps      []*Step
	Endpoints  []*Endpoint  // every route: resources expanded by lower, then Custom
	Pages      []*Page      // frontend only
	Components []*Component // frontend only
	Look       *Look        // frontend only: how the app looks, for the model that writes its theme
}

// Look is a frontend's look line.
type Look struct {
	Text string
	Pos  diag.Pos
}

// Page is a frontend route that loads records and shows components in
// order, or a form that submits to an endpoint.
type Page struct {
	Route    string // params as [id]
	Load     string // the name of what it loads; empty for none, and for forms
	Resource string // what it loads or submits to, in the backend the frontend uses; empty for a form on a custom endpoint
	Entity   string
	ByID     bool // loads the record named by the route's [id]; otherwise lists them all
	Sections []*Section
	Empty    string // pages that load a list: shown when there are no records
	NotFound string // by-id pages: shown when the record doesn't exist
	Form     string // form pages: the endpoint it submits to, like horses.create or swipe
	Then     string // form pages: where to go after a save; [id] is the saved record's
}

// Section is one component a page shows, full width, in order.
type Section struct {
	Component string
	List      bool // the page's list, each record as one Component in a grid
	Arg       bool // the component gets what the page loads
}

// Component shows nothing, one record or a list. What it shows is also what
// it reads, so a page loads exactly that.
type Component struct {
	Name      string
	Pos       diag.Pos
	Look      string   // how it should look, for the model that writes its body
	Says      []string // its words, from the spec, in order
	Param     string   // empty for a component that takes no record
	Entity    string
	List      bool // the param is a list of Entity
	Shows     []*Show
	LinksTo   string   // the route of the page it links to, empty for none
	LinkBy    []string // the param's field for each of that route's params, in order
	LinkLabel string   // the link's button text; empty makes the whole component the link
	Actions   []*Action
}

// ActionKind is what an action under a component does.
type ActionKind int

const (
	Open   ActionKind = iota // go to a page
	Remove                   // delete the record
)

// Action is a link or button that a template renders beside a component's
// body, so the model never sees it (D-020).
type Action struct {
	Kind     ActionKind
	Label    string
	Route    string   // Open: the page
	RouteBy  []string // Open: the record's field for each of Route's params
	Endpoint string   // Remove: the endpoint, like horses.delete
	Then     string   // where to go afterwards; empty to stay
	ThenBy   []string // the record's field for each of Then's params
}

// Show is one field or relation a component shows.
type Show struct {
	Name     string
	Field    *Field    // set for a field
	Relation *Relation // set for a relation; a has-many one shows as chips
	Display  string    // for a relation: the target's display field
	Resource string    // for a relation: the backend resource that loads it
	Href     string    // for a relation: the page each related record links to, params as [id]; empty for none (D-024)
}

// Op is a standard operation on a resource.
type Op int

const (
	List Op = iota
	Get
	Create
	Update
	Delete
)

var OpNames = [...]string{"list", "get", "create", "update", "delete"}

func (o Op) String() string { return OpNames[o] }

type Resource struct {
	Name   string
	Entity string
	Ops    []Op // in Op order
}

// Endpoint is one HTTP route.
type Endpoint struct {
	Name     string // e.g. products.list
	Method   string
	Route    string // params as [name]
	Resource string
	Op       Op
	Entity   string
	Input    []*Input // nil when the endpoint takes no body
	Partial  bool     // every input is optional (updates)
	Output   string   // entity name, empty when nothing is returned
	List     bool     // output is a list
	Errors   []string
	Relation string // for a relation read: the has-many relation it lists
	Flow     *Flow  // set for custom endpoints
}

// Flow is what a custom endpoint does: check its input, then run its steps
// in order. Step parameters are bound by name to an input or an earlier result.
type Flow struct {
	Input []*Value
	Steps []*Step
}

// Value is an input, step parameter or step result: a scalar, or a record
// of an entity.
type Value struct {
	Name   string
	Field  *Field // the scalar's type; nil for a record
	Entity string // the record's entity
}

// Step is a function whose body a model writes.
type Step struct {
	Name   string
	Pos    diag.Pos
	Params []*Value
	Result *Value // nil when the step returns nothing
	Fails  []string
	Reads  []string // entities whose tables it only reads
	Writes []string // entities whose tables it writes, and may read
	Intent string
}

// Input is one field of a request body: an entity field, or a foreign key
// sent as the referenced entity's id.
type Input struct {
	Name  string
	Field *Field // nil for a foreign key
	Ref   string // entity a foreign key points to
}

func (a *App) Entity(name string) *Entity {
	for _, e := range a.Entities {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func (a *App) Enum(name string) *Enum {
	for _, e := range a.Enums {
		if e.Name == name {
			return e
		}
	}
	return nil
}
