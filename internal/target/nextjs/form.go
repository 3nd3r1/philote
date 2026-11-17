package nextjs

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// formInput is one field a form sends: a scalar, or a record picked by id.
type formInput struct {
	Name  string
	Field *ir.Field // nil for a record
	Ref   string    // the record's entity
}

// controlView is one form control. Label is a JavaScript string literal;
// Attrs are extra JSX attributes, each with a leading space; Read and
// Initial are TypeScript expressions.
type controlView struct {
	Name, Label, Control, Type, Attrs string
	Choices                           []string // an enum's values
	Read, Initial                     string
}

// optionView is a list of records a select offers, labelled by display.
type optionView struct{ Name, Client, Label string }

// formView is everything a form page's three templates print.
type formView struct {
	Route, Title, NotFound, Call, Then string
	ByID                               bool
	Client                             string // loads the record an edit form starts from
	PageClients, ActionClients         string // imports from @/lib/api
	Readers, Types                     string // imports from @/lib/form and @/lib/types
	LocalDateTime                      bool
	UIImports                          string // only the controls the form renders
	Fields                             []controlView
	Options                            []optionView
}

// formFiles renders a form page: the page, its client form, and the
// server action it posts to (D-020).
func formFiles(app *ir.App, backend *ir.Part, pg *ir.Page) []target.File {
	v := newFormView(app, backend, pg)
	dir := "src/app" + pg.Route + "/"
	files := []target.File{
		render("form-page.tsx.tmpl", dir+"page.tsx", v),
		render("form.tsx.tmpl", dir+"form.tsx", v),
		render("form-actions.ts.tmpl", dir+"actions.ts", v),
	}
	if pg.ByID {
		files = append(files, render("not-found.tsx.tmpl", dir+"not-found.tsx", pageView{Route: pg.Route, NotFound: jsString(pg.NotFound)}))
	}
	return files
}

func newFormView(app *ir.App, backend *ir.Part, pg *ir.Page) formView {
	ep := endpoint(backend, pg.Form)
	v := formView{Route: pg.Route, Title: jsString(formTitle(pg)), NotFound: jsString(pg.NotFound), ByID: pg.ByID}
	pageClients, actionClients := map[string]bool{}, map[string]bool{}
	readers, types, uses := map[string]bool{"failed": true}, map[string]bool{}, map[string]bool{}
	if pg.Resource != "" {
		client := camel(pg.Resource)
		v.Call, actionClients[client] = client+"."+strings.TrimPrefix(pg.Form, pg.Resource+"."), true
		if pg.ByID {
			v.Client, pageClients[client] = client, true
		}
	} else {
		v.Call, actionClients[camel(ep.Name)] = camel(ep.Name), true
	}
	for _, in := range formInputs(ep) {
		f := newFieldView(app, in, readers, types)
		if in.Field != nil && in.Field.Kind == ir.DateTime && pg.ByID {
			v.LocalDateTime = true
		}
		if in.Ref != "" {
			ent := app.Entity(in.Ref)
			client := camel(lister(backend, in.Ref))
			pageClients[client] = true
			v.Options = append(v.Options, optionView{in.Name, client, "r." + ent.Display + optionalDefault(findField(ent, ent.Display))})
		}
		v.Fields, uses[f.Control] = append(v.Fields, f), true
	}
	v.Then = jsString(pg.Then)
	if strings.Contains(pg.Then, "[id]") {
		v.Then = "`" + strings.Replace(pg.Then, "[id]", "${encodeURIComponent(result.value.id)}", 1) + "`"
	}
	v.PageClients = strings.Join(sortedKeys(pageClients), ", ")
	v.ActionClients = strings.Join(sortedKeys(actionClients), ", ")
	v.Readers = strings.Join(sortedKeys(readers), ", ")
	v.Types = imports(app, types)
	v.UIImports = uiImports(uses)
	return v
}

// uiImports imports the base's controls a form uses, in import order.
func uiImports(uses map[string]bool) string {
	var lines []string
	if uses["input"] {
		lines = append(lines, `import { Input } from "@/components/ui/input";`)
	}
	if uses["select"] || uses["records"] {
		lines = append(lines, `import { NativeSelect, NativeSelectOption } from "@/components/ui/native-select";`)
	}
	if uses["textarea"] {
		lines = append(lines, `import { Textarea } from "@/components/ui/textarea";`)
	}
	return strings.Join(lines, "\n")
}

// formInputs are an endpoint's inputs: a resource's fields and foreign
// keys, or a custom endpoint's values.
func formInputs(ep *ir.Endpoint) []formInput {
	var out []formInput
	if ep.Flow != nil {
		for _, val := range ep.Flow.Input {
			out = append(out, formInput{val.Name, val.Field, val.Entity})
		}
		return out
	}
	for _, in := range ep.Input {
		out = append(out, formInput{in.Name, in.Field, in.Ref})
	}
	return out
}

func newFieldView(app *ir.App, in formInput, readers, types map[string]bool) controlView {
	v := controlView{Name: in.Name, Control: "input", Type: "text"}
	f := in.Field
	if f == nil {
		v.Label = jsString(words(strings.TrimSuffix(in.Name, "_id")))
		v.Control, v.Attrs = "records", " required"
		v.Read, v.Initial = read("text", in.Name), "record."+in.Name
		readers["text"] = true
		return v
	}
	v.Label = jsString(words(in.Name))
	if !f.Optional && f.Kind != ir.Bool {
		v.Attrs = " required"
	}
	reader := "text"
	switch f.Kind {
	case ir.String, ir.Markdown:
		if f.Kind == ir.Markdown || f.Max == nil || *f.Max > 160 {
			v.Control = "textarea"
		}
		v.Attrs += bounds("minLength", "maxLength", f)
	case ir.Email:
		v.Type = "email"
	case ir.Int:
		v.Type, reader = "number", "int"
		v.Attrs += bounds("min", "max", f)
	case ir.Bool:
		v.Control, reader = "checkbox", "bool"
	case ir.Date:
		v.Type = "date"
	case ir.DateTime:
		v.Type, reader = "datetime-local", "dateTime"
	case ir.EnumValue:
		v.Control, v.Choices = "select", app.Enum(f.Enum).Values
		types[f.Enum] = true
	}
	readers[reader] = true
	v.Read = read(reader, in.Name)
	if f.Kind == ir.EnumValue {
		v.Read = fmt.Sprintf("text<%s>(data, %q)", f.Enum, in.Name)
	}
	v.Initial = initial(f)
	return v
}

func read(reader, name string) string { return fmt.Sprintf("%s(data, %q)", reader, name) }

// initial is a field's value in an edit form, as the string its control holds.
func initial(f *ir.Field) string {
	value := "record." + f.Name
	switch {
	case f.Kind == ir.Bool:
		return value + ` ? "on" : ""`
	case f.Kind == ir.Int && f.Optional:
		return value + ` === undefined ? "" : String(` + value + ")"
	case f.Kind == ir.Int:
		return "String(" + value + ")"
	case f.Kind == ir.DateTime && f.Optional:
		return value + " ? localDateTime(" + value + `) : ""`
	case f.Kind == ir.DateTime:
		return "localDateTime(" + value + ")"
	}
	return value + optionalDefault(f)
}

func optionalDefault(f *ir.Field) string {
	if f.Optional {
		return ` ?? ""`
	}
	return ""
}

func bounds(lo, hi string, f *ir.Field) string {
	var out string
	if f.Min != nil {
		out += " " + lo + "={" + strconv.Itoa(*f.Min) + "}"
	}
	if f.Max != nil {
		out += " " + hi + "={" + strconv.Itoa(*f.Max) + "}"
	}
	return out
}

// formTitle is "New Horse", "Edit Horse", or a custom endpoint's name in words.
func formTitle(pg *ir.Page) string {
	switch {
	case pg.Resource == "":
		return words(pg.Form)
	case pg.ByID:
		return "Edit " + pg.Entity
	}
	return "New " + pg.Entity
}

// words turns matched_at into "Matched at".
func words(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	return strings.ToUpper(s[:1]) + s[1:]
}

// lister is the one resource that lists entity. check guarantees it.
func lister(backend *ir.Part, entity string) string {
	for _, r := range backend.Resources {
		if r.Entity == entity && slices.Contains(r.Ops, ir.List) {
			return r.Name
		}
	}
	panic("nextjs: no resource lists " + entity)
}

func endpoint(backend *ir.Part, name string) *ir.Endpoint {
	for _, ep := range backend.Endpoints {
		if ep.Name == name {
			return ep
		}
	}
	panic("nextjs: no endpoint " + name)
}
