package hono

import (
	"bytes"
	"embed"
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
func (Hono) Version() string   { return "0.1.0" }

func (Hono) Generate(app *ir.App, part *ir.Part) ([]target.File, error) {
	files, err := baseFiles()
	if err != nil {
		return nil, err
	}
	pkg := struct{ Name string }{app.Name + "-" + part.Name}
	appView := struct{ Resources []resourceView }{}
	var rendered []target.File
	for _, r := range part.Resources {
		v := newResourceView(app, part, r)
		appView.Resources = append(appView.Resources, v)
		rendered = append(rendered, render("resource.ts.tmpl", "src/resources/"+r.Name+".ts", v))
	}
	rendered = append(rendered,
		render("package.json.tmpl", "package.json", pkg),
		render("app.ts.tmpl", "src/app.ts", appView),
	)
	files = append(files, rendered...)
	slices.SortFunc(files, func(a, b target.File) int { return strings.Compare(a.Path, b.Path) })
	return files, nil
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
