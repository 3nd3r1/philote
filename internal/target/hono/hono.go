package hono

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"text/template"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

//go:embed all:base
var base embed.FS

//go:embed templates/*.tmpl
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))

// Hono builds a backend part as a Hono server on Node.
type Hono struct{}

func (Hono) Name() string      { return "hono" }
func (Hono) Kind() ir.PartKind { return ir.Backend }
func (Hono) Version() string   { return "0.3.0" }

// reserved are names the generated endpoint handlers use themselves.
var reserved = []string{"ctx", "checked", "error", "app"}

func (Hono) Generate(app *ir.App, part *ir.Part) (target.Output, error) {
	if err := checkNames(part); err != nil {
		return target.Output{}, err
	}
	files, err := baseFiles()
	if err != nil {
		return target.Output{}, err
	}
	types := render("types.ts.tmpl", "src/types.ts", newTypesView(app))
	out := target.Output{Files: append(files, types,
		render("package.json.tmpl", "package.json", struct{ Name string }{app.Name + "-" + part.Name}))}

	appView := struct{ Resources, Endpoints []mountView }{}
	for _, r := range part.Resources {
		v := newResourceView(app, part, r)
		appView.Resources = append(appView.Resources, mountView{v.Func, v.File})
		out.Files = append(out.Files, render("resource.ts.tmpl", "src/resources/"+r.Name+".ts", v))
	}
	for _, ep := range part.Custom {
		v := newEndpointView(app, ep)
		appView.Endpoints = append(appView.Endpoints, mountView{v.Func, v.Name})
		out.Files = append(out.Files, render("endpoint.ts.tmpl", "src/endpoints/"+ep.Name+".ts", v))
	}
	for _, s := range part.Steps {
		v := newStepView(app, s)
		path := "src/steps/" + s.Name + ".ts"
		out.Files = append(out.Files, render("step.ts.tmpl", path, v))
		out.Slots = append(out.Slots, target.Slot{
			Node:   part.Name + "/" + s.Name,
			File:   path,
			Marker: marker,
			Prompt: prompt(v, string(types.Data)),
		})
	}
	out.Files = append(out.Files, render("app.ts.tmpl", "src/app.ts", appView))
	slices.SortFunc(out.Files, func(a, b target.File) int { return strings.Compare(a.Path, b.Path) })
	return out, nil
}

type mountView struct{ Func, File string }

func checkNames(part *ir.Part) error {
	for _, ep := range part.Custom {
		for _, v := range ep.Flow.Input {
			if slices.Contains(reserved, v.Name) {
				return fmt.Errorf("endpoint %s: the input name %q clashes with the generated handler; rename it", ep.Name, v.Name)
			}
		}
		for _, s := range ep.Flow.Steps {
			if s.Result != nil && slices.Contains(reserved, s.Result.Name) {
				return fmt.Errorf("step %s: the result name %q clashes with the generated handler; rename it", s.Name, s.Result.Name)
			}
		}
	}
	return nil
}

// baseFiles are copied as they are.
func baseFiles() ([]target.File, error) {
	var out []target.File
	err := fs.WalkDir(base, "base", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := base.ReadFile(path)
		out = append(out, target.File{Path: strings.TrimPrefix(path, "base/"), Data: data})
		return err
	})
	return out, err
}

// render executes a template. The views are built from checked IR, so a
// failure here is a bug in the target, not in the user's spec.
func render(name, path string, view any) target.File {
	var b bytes.Buffer
	if err := templates.ExecuteTemplate(&b, name, view); err != nil {
		panic("hono: template " + name + ": " + err.Error())
	}
	return target.File{Path: path, Data: b.Bytes()}
}
