package nextjs

import (
	"fmt"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// reserved are names the generated component and read code use itself.
var reserved = []string{"props", "item", "must", "cn", "Icons"}

// marker is the line in a component file that the model's body replaces.
const marker = "    // philote:body"

type componentView struct{ Name, Param, Link, Marker string }

type relationView struct{ Name, Type string }

type readView struct {
	Name, Param, Entity, Pick, LoadNames string
	Relations                            []relationView
	Loads, Values                        []string
}

type readsView struct {
	Clients, Types string
	Components     []readView
}

// components renders each component around a slot for its body, and
// reads.ts, which loads their props.
func components(app *ir.App, part *ir.Part) ([]target.File, []target.Slot, error) {
	var files []target.File
	var slots []target.Slot
	reads := readsView{}
	clients, types := map[string]bool{}, map[string]bool{}
	for _, c := range part.Components {
		if err := checkComponent(c); err != nil {
			return nil, nil, err
		}
		path := "src/components/common/" + kebab(c.Name) + ".tsx"
		files = append(files, render("component.tsx.tmpl", path, newComponentView(c)))
		slots = append(slots, target.Slot{
			Node:   part.Name + "/" + c.Name,
			File:   path,
			Marker: marker,
			Prompt: componentPrompt(app, part, c),
		})
		reads.Components = append(reads.Components, newReadView(c, clients, types))
	}
	reads.Clients = strings.Join(sortedKeys(clients), ", ")
	reads.Types = imports(app, types)
	return append(files, render("reads.ts.tmpl", "src/lib/reads.ts", reads)), slots, nil
}

func checkComponent(c *ir.Component) error {
	if slices.Contains(reserved, c.Param) {
		return fmt.Errorf("component %s: the parameter name %q clashes with the generated code; rename it", c.Name, c.Param)
	}
	for _, s := range c.Shows {
		if s.Relation != nil && s.Name == c.Param {
			return fmt.Errorf("component %s: the relation %s has the parameter's name; rename the parameter", c.Name, s.Name)
		}
	}
	return nil
}

func newComponentView(c *ir.Component) componentView {
	v := componentView{Name: c.Name, Param: c.Param, Marker: marker}
	if c.LinksTo != "" {
		link := c.LinksTo
		for _, f := range c.LinkBy {
			link = strings.Replace(link, "["+f+"]", "${encodeURIComponent(props."+c.Param+"."+f+")}", 1)
		}
		v.Link = "`" + link + "`"
	}
	return v
}

// newReadView builds a component's props type and the read that fills it:
// the record's shown fields, plus one request per relation it shows.
func newReadView(c *ir.Component, clients, types map[string]bool) readView {
	v := readView{Name: c.Name, Param: c.Param, Entity: c.Entity}
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
		if rel.Many {
			v.Relations = append(v.Relations, relationView{s.Name, pick + "[]"})
			v.Loads = append(v.Loads, fmt.Sprintf("%s.%s(%s.id)", client, camel(rel.Name), c.Param))
			v.Values = append(v.Values, fmt.Sprintf("%s: %s.map((item) => (%s))", s.Name, s.Name, fmt.Sprintf(record, "item")))
		} else {
			v.Relations = append(v.Relations, relationView{s.Name, pick})
			v.Loads = append(v.Loads, fmt.Sprintf("%s.get(%s.%s)", client, c.Param, rel.FK))
			v.Values = append(v.Values, s.Name+": "+fmt.Sprintf(record, s.Name))
		}
		names = append(names, s.Name)
	}
	fieldValues := make([]string, len(fields))
	for i, f := range fields {
		fieldValues[i] = f + ": " + c.Param + "." + f
	}
	v.Values = append(fieldValues, v.Values...)
	v.Pick, v.LoadNames = union(fields), strings.Join(names, ", ")
	return v
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
