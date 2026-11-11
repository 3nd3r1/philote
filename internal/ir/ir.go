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
	Name      string
	Kind      PartKind
	Target    string
	Uses      string // for a frontend: the backend part it calls
	Resources []*Resource
	Custom    []*Endpoint // endpoints declared with steps
	Steps     []*Step
	Endpoints []*Endpoint // every route: resources expanded by lower, then Custom
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
