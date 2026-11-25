package nextjs

import (
	"fmt"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// reserved are names the generated component and read code use itself.
var reserved = []string{"props", "item", "record", "must", "cn", "link", "actions", "says",
	"useState", "toast", "chip", "current"}

// imported are the names every component file may import, which a component
// can't take as its own name.
var imported = []string{"Link", "Badge", "Button", "buttonVariants", "Card", "CardContent",
	"CardFooter", "CardHeader", "CardTitle", "Mark", "FormField", "SubmitButton"}

type componentView struct {
	Name, Kind, Param    string
	Client, Props        bool
	LinkExpr             string
	LinkLabel            string
	Actions              string
	Pick                 string
	Call, CallFile       string
	CallRecordArg        string
	CallBoolArg          string
	MatchField           string
	Says                 []string
	Title                *showView
	AvatarExpr           string
	AllFields            []showView
	Fields               []showView
	MetaFields           []showView
	LongFields           []showView
	ChipFields           []showView
	FormatImports        string
	ImportLink           bool
	ImportMark           bool
	ImportBadge          bool
	ImportCard           bool
	ImportCardContent    bool
	ImportCardFooter     bool
	ImportButton         bool
	ImportButtonVariants bool
}

type relationView struct{ Name, Type string }

type readView struct {
	Name, Param, Entity, Pick, LoadNames string
	List                                 bool // reads a list, each record as record
	Relations                            []relationView
	Loads, Values                        []string
}

type readsView struct {
	Clients, Types string
	Components     []readView
}

// components renders each component from its built-in kind template, and
// reads.ts, which loads their props. No frontend model calls are made.
func components(app *ir.App, backend, part *ir.Part) ([]target.File, error) {
	var files []target.File
	reads := readsView{}
	clients, types := map[string]bool{}, map[string]bool{}
	for _, c := range part.Components {
		if err := checkComponent(c); err != nil {
			return nil, err
		}
		path := "src/components/common/" + kebab(c.Name) + ".tsx"
		files = append(files, render("component.tsx.tmpl", path, newComponentView(app, backend, c)))
		files = append(files, actionFiles(app, part, c)...)
		files = append(files, pickFiles(c)...)
		files = append(files, callFiles(app, backend, c)...)
		if c.Param != "" {
			reads.Components = append(reads.Components, newReadView(c, clients, types))
		}
	}
	reads.Clients = strings.Join(sortedKeys(clients), ", ")
	reads.Types = imports(app, types)
	return append(files, render("reads.ts.tmpl", "src/lib/reads.ts", reads)), nil
}

func checkComponent(c *ir.Component) error {
	if slices.Contains(imported, c.Name) {
		return fmt.Errorf("component %s: the name clashes with a component the app already has; rename it", c.Name)
	}
	if slices.Contains(reserved, c.Param) {
		return fmt.Errorf("component %s: the parameter name %q clashes with the generated code; rename it", c.Name, c.Param)
	}
	for _, s := range c.Shows {
		if s.Relation != nil && s.Name == c.Param {
			return fmt.Errorf("component %s: the relation %s has the parameter's name; rename the parameter", c.Name, s.Name)
		}
	}
	if c.Calls != nil {
		if name := camel(c.Calls.Endpoint); name == c.Param || slices.Contains(reserved, name) {
			return fmt.Errorf("component %s: the action %s clashes with the generated code; rename the endpoint or the parameter", c.Name, name)
		}
	}
	return nil
}

func newComponentView(app *ir.App, backend *ir.Part, c *ir.Component) componentView {
	v := componentView{
		Name:  c.Name,
		Kind:  c.Kind,
		Param: c.Param,
		Props: c.Param != "",
		Says:  jsStrings(c.Says),
	}
	if c.LinksTo != "" {
		v.LinkExpr = linkExpr(c)
		v.LinkLabel = c.LinkLabel
	}
	if len(c.Actions) > 0 {
		v.Actions = kebab(c.Name) + "-actions"
	}
	if c.Picks != "" {
		v.Pick = kebab(c.Name) + "-pick"
	}
	setComponentCalls(backend, c, &v)
	buildComponentFields(app, c, &v)
	setComponentImports(c, &v)
	return v
}

func linkExpr(c *ir.Component) string {
	if len(c.LinkBy) == 0 {
		return jsString(c.LinksTo)
	}
	link := c.LinksTo
	for _, f := range c.LinkBy {
		link = strings.Replace(link, "["+f+"]", "${encodeURIComponent("+c.Param+"."+f+")}", 1)
	}
	return "`" + link + "`"
}

func jsStrings(items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = jsString(s)
	}
	return out
}

// newReadView builds a component's props type and the read that fills it:
// the record's shown fields, plus one request per relation it shows.
func newReadView(c *ir.Component, clients, types map[string]bool) readView {
	v := readView{Name: c.Name, Param: c.Param, Entity: c.Entity, List: c.List}
	// A list's reads run once per record, named record.
	name := c.Param
	if c.List {
		name = "record"
	}
	types[c.Entity] = true
	fields := []string{"id"}
	for _, s := range c.Shows {
		if s.Field != nil {
			fields = append(fields, s.Name)
		}
	}
	for _, f := range c.LinkBy {
		if !slices.Contains(fields, f) {
			fields = append(fields, f)
		}
	}
	var names []string
	for _, s := range c.Shows {
		if s.Relation == nil {
			continue
		}
		rel, client := s.Relation, camel(s.Resource)
		clients[client], types[rel.Target] = true, true
		pick := fmt.Sprintf("Pick<%s, %s>", rel.Target, union([]string{"id", s.Display}))
		record := fmt.Sprintf("{ id: %%[1]s.id, %[1]s: %%[1]s.%[1]s }", s.Display)
		if s.Href != "" {
			pick += " & { href: string }"
			href := strings.Replace(s.Href, "[id]", "${encodeURIComponent(%[1]s.id)}", 1)
			record = fmt.Sprintf("{ id: %%[1]s.id, %[1]s: %%[1]s.%[1]s, href: `%[2]s` }", s.Display, href)
		}
		if rel.Many {
			v.Relations = append(v.Relations, relationView{s.Name, list(pick)})
			v.Loads = append(v.Loads, fmt.Sprintf("%s.%s(%s.id)", client, camel(rel.Name), name))
			v.Values = append(v.Values, fmt.Sprintf("%s: %s.map((item) => (%s))", s.Name, s.Name, fmt.Sprintf(record, "item")))
		} else {
			v.Relations = append(v.Relations, relationView{s.Name, pick})
			v.Loads = append(v.Loads, fmt.Sprintf("%s.get(%s.%s)", client, name, rel.FK))
			v.Values = append(v.Values, s.Name+": "+fmt.Sprintf(record, s.Name))
		}
		names = append(names, s.Name)
	}
	fieldValues := make([]string, len(fields))
	for i, f := range fields {
		fieldValues[i] = f + ": " + name + "." + f
	}
	v.Values = append(fieldValues, v.Values...)
	v.Pick, v.LoadNames = union(fields), strings.Join(names, ", ")
	return v
}

// list makes a record type into an array type, with parentheses when it's an
// intersection.
func list(t string) string {
	if strings.Contains(t, " & ") {
		return "(" + t + ")[]"
	}
	return t + "[]"
}

// union prints names as a union of string literal types.
func union(names []string) string {
	return strings.Join(quoted(names), " | ")
}

func sortedKeys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
