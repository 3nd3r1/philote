package nextjs

import (
	"fmt"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// reserved are names the generated component and read code use itself,
// and the hooks and toast a calling component imports.
var reserved = []string{"props", "item", "record", "must", "cn", "Icons", "link", "actions", "says",
	"useEffect", "useRef", "useState", "useTransition", "toast"}

// imported are the names every component file imports, which a component
// can't take as its own name.
var imported = []string{"Link", "Badge", "Card", "CardAction", "CardContent", "CardDescription",
	"CardFooter", "CardHeader", "CardTitle", "Separator", "Mark", "FormField", "SubmitButton"}

// marker is the line in a component file that the model's body replaces.
const marker = "    // philote:body"

type componentView struct {
	Name, Marker string
	Client       bool   // the component calls an endpoint, so it runs in the browser
	Call         string // the action the body calls, empty for none
	CallFile     string // the file that action is in
	Props        bool   // the component takes a record or a list, as XProps
	Extras       string // props the wrapper adds for the body: link and actions
	Says         string // the says array's items, as JavaScript strings
	BodyArgs     string // the body's parameter list
	Actions      string // the actions file the body places, empty for none
	Pick         string // the pick button the body wraps each record in, empty for none
	ImportLink   bool   // the body renders a Link
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

// components renders each component around a slot for its body, and
// reads.ts, which loads their props.
func components(app *ir.App, backend, part *ir.Part) ([]target.File, []target.Slot, error) {
	var files []target.File
	var slots []target.Slot
	reads := readsView{}
	clients, types := map[string]bool{}, map[string]bool{}
	for _, c := range part.Components {
		if err := checkComponent(c); err != nil {
			return nil, nil, err
		}
		path := "src/components/common/" + kebab(c.Name) + ".tsx"
		files = append(files, render("component.tsx.tmpl", path, newComponentView(app, c)))
		files = append(files, actionFiles(app, part, c)...)
		files = append(files, pickFiles(c)...)
		files = append(files, callFiles(app, backend, c)...)
		slots = append(slots, target.Slot{
			Node:   part.Name + "/" + c.Name,
			File:   path,
			Marker: marker,
			Prompt: componentPrompt(app, part, c),
		})
		if c.Param != "" {
			reads.Components = append(reads.Components, newReadView(c, clients, types))
		}
	}
	reads.Clients = strings.Join(sortedKeys(clients), ", ")
	reads.Types = imports(app, types)
	return append(files, render("reads.ts.tmpl", "src/lib/reads.ts", reads)), slots, nil
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

// newComponentView builds the wrapper that hands the body its props, plus
// the link and actions it places, which the model never sees the targets of.
func newComponentView(app *ir.App, c *ir.Component) componentView {
	v := componentView{Name: c.Name, Marker: marker, Props: c.Param != "", Client: c.Calls != nil}
	var extras []string
	if c.LinksTo != "" {
		link := jsString(c.LinksTo)
		if len(c.LinkBy) > 0 {
			link = c.LinksTo
			for _, f := range c.LinkBy {
				link = strings.Replace(link, "["+f+"]", "${encodeURIComponent(props."+c.Param+"."+f+")}", 1)
			}
			link = "`" + link + "`"
		}
		if c.LinkLabel != "" {
			link += ", label: " + jsString(c.LinkLabel)
		}
		extras = append(extras, fmt.Sprintf(" link={{ href: %s }}", link))
	}
	if len(c.Actions) > 0 {
		v.Actions = kebab(c.Name) + "-actions"
		extras = append(extras, fmt.Sprintf(" actions={<%sActions id={props.%s.id} />}", c.Name, c.Param))
	}
	if c.Calls != nil {
		v.Call, v.CallFile = camel(c.Calls.Endpoint), kebab(c.Name)+"-call"
		extras = append(extras, fmt.Sprintf(" %[1]s={%[1]s}", v.Call))
	}
	v.Extras = strings.Join(extras, "")
	if c.Picks != "" {
		v.Pick = kebab(c.Name) + "-pick"
	}
	if len(c.Says) > 0 {
		v.Says = strings.Join(jsStrings(c.Says), ", ")
	}
	names, types := bodyParams(app, c)
	v.BodyArgs = bodyArgs(c, names, types)
	v.ImportLink = c.LinksTo != "" || slices.ContainsFunc(c.Shows, func(s *ir.Show) bool { return s.Href != "" })
	return v
}

// bodyParams names what the body receives, and the type of what the
// wrapper adds beside the record: the link, the actions and the call.
func bodyParams(app *ir.App, c *ir.Component) ([]string, string) {
	var names, types []string
	if c.Param != "" {
		names = append(names, c.Param)
	}
	if c.LinksTo != "" {
		names = append(names, "link")
		if c.LinkLabel != "" {
			types = append(types, "link: { href: string; label: string }")
		} else {
			types = append(types, "link: { href: string }")
		}
	}
	if len(c.Actions) > 0 {
		names = append(names, "actions")
		types = append(types, "actions: React.ReactNode")
	}
	if c.Calls != nil {
		name := camel(c.Calls.Endpoint)
		names = append(names, name)
		types = append(types, name+": "+actionType(app, c.Calls))
	}
	return names, strings.Join(types, "; ")
}

// bodyArgs is the body's parameter list: its props, plus extras like
// `link: { href: string }`.
func bodyArgs(c *ir.Component, names []string, extras string) string {
	if len(names) == 0 {
		return ""
	}
	typ := c.Name + "Props"
	switch {
	case c.Param == "":
		typ = "{ " + extras + " }"
	case extras != "":
		typ += " & { " + extras + " }"
	}
	return "{ " + strings.Join(names, ", ") + " }: " + typ
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
