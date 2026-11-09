package ir

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
	Resources []*Resource
	Endpoints []*Endpoint // filled by lower
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
