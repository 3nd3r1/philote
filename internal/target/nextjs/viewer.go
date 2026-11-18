package nextjs

import (
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// viewerCookie holds the viewer's id.
const viewerCookie = "philote-viewer"

type viewerView struct {
	Entity, Client, Cookie string
	Picker, PickerJS       string   // empty when no page picks the viewer
	Matcher                []string // the routes that show the viewer, as proxy matchers
}

type pickView struct {
	Component, File, Entity string
	Then, ThenJS            string
}

// viewerFiles read the viewer from its cookie. proxy.ts sends a visitor
// without one to the picker before a page that shows the viewer renders, so
// that's a real redirect; check guarantees such a page has a picker.
func viewerFiles(part *ir.Part) []target.File {
	v := part.Viewer
	if v == nil {
		return nil
	}
	view := viewerView{Entity: v.Entity, Client: camel(v.Resource), Cookie: viewerCookie, Picker: v.Picker}
	if v.Picker != "" {
		view.PickerJS = jsString(v.Picker)
	}
	for _, pg := range part.Pages {
		if pg.Viewer {
			view.Matcher = append(view.Matcher, jsString(strings.ReplaceAll(pg.Route, "[id]", ":id")))
		}
	}
	files := []target.File{render("viewer.ts.tmpl", "src/lib/viewer.ts", view)}
	if len(view.Matcher) > 0 {
		files = append(files, render("proxy.ts.tmpl", "src/proxy.ts", view))
	}
	return files
}

// pickFiles render the button a picking component's body wraps each record
// in, and the server action behind it. The model never sees where picking
// goes, so changing then doesn't touch the pinned body.
func pickFiles(c *ir.Component) []target.File {
	if c.Picks == "" {
		return nil
	}
	v := pickView{Component: c.Name, File: kebab(c.Name), Entity: c.Entity, Then: c.Picks, ThenJS: jsString(c.Picks)}
	return []target.File{
		render("pick.ts.tmpl", "src/lib/actions/"+v.File+"-pick.ts", v),
		render("pick.tsx.tmpl", "src/components/common/"+v.File+"-pick.tsx", v),
	}
}
