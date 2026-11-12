package nextjs

import (
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// actionView is one button or link under a component's body. Label,
// Href, Then and Confirm are TypeScript expressions.
type actionView struct {
	Open                 bool
	Label, Href, Confirm string
}

// actionsView is a component's actions file and the server action its
// delete calls.
type actionsView struct {
	Component, File, Client, Then string
	Buttons                       string // imports from the base's button
	Open, Delete                  bool
	Actions                       []actionView
}

// actionFiles renders the buttons under a component's body, and the
// server action behind its delete. The model never sees these, so changing
// an action doesn't touch the component's pinned body (D-020).
func actionFiles(part *ir.Part, c *ir.Component) []target.File {
	if len(c.Actions) == 0 {
		return nil
	}
	v := actionsView{Component: c.Name, File: kebab(c.Name)}
	for _, a := range c.Actions {
		if a.Kind == ir.Open {
			label := a.Label
			if label == "" {
				label = pageTitle(findPage(part, a.Route))
			}
			v.Open = true
			v.Actions = append(v.Actions, actionView{Open: true, Label: jsString(label), Href: recordHref(a.Route)})
			continue
		}
		label := a.Label
		if label == "" {
			label = "Delete"
		}
		v.Delete, v.Client = true, camel(strings.TrimSuffix(a.Endpoint, ".delete"))
		v.Then = recordHref(a.Then)
		v.Actions = append(v.Actions, actionView{Label: jsString(label), Confirm: jsString("Delete this " + strings.ToLower(c.Entity) + "?")})
	}
	var buttons []string
	if v.Delete {
		buttons = append(buttons, "Button")
	}
	if v.Open {
		buttons = append(buttons, "buttonVariants")
	}
	v.Buttons = strings.Join(buttons, ", ")
	files := []target.File{render("actions.tsx.tmpl", "src/components/common/"+v.File+"-actions.tsx", v)}
	if v.Delete {
		files = append(files, render("remove.ts.tmpl", "src/lib/actions/"+v.File+".ts", v))
	}
	return files
}

// recordHref fills a route's [id] with the record's id. Routes only have
// [id] as a param (check).
func recordHref(route string) string {
	if route == "" {
		return ""
	}
	if !strings.Contains(route, "[id]") {
		return jsString(route)
	}
	return "`" + strings.Replace(route, "[id]", "${encodeURIComponent(id)}", 1) + "`"
}

func findPage(part *ir.Part, route string) *ir.Page {
	for _, pg := range part.Pages {
		if pg.Route == route {
			return pg
		}
	}
	panic("nextjs: no page " + route)
}
