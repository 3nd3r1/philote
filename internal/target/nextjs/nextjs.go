package nextjs

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

// Nextjs builds a frontend part as a Next.js app with a typed client for
// the backend it uses.
type Nextjs struct{}

func (Nextjs) Name() string      { return "nextjs" }
func (Nextjs) Kind() ir.PartKind { return ir.Frontend }
func (Nextjs) Version() string   { return "0.2.0" }

func (Nextjs) Generate(app *ir.App, part *ir.Part) (target.Output, error) {
	backend := usedBackend(app, part)
	api, err := newAPIView(app, backend)
	if err != nil {
		return target.Output{}, err
	}
	views, err := components(app, part)
	if err != nil {
		return target.Output{}, err
	}
	generated := slices.Concat(views, []target.File{
		render("config.ts.tmpl", "src/lib/config.ts", struct{ Name, Description string }{
			jsString(app.Name), jsString("The " + app.Name + " app, built by philote."),
		}),
		render("types.ts.tmpl", "src/lib/types.ts", newTypesView(app)),
		render("api.ts.tmpl", "src/lib/api.ts", api),
	}, pages(part))
	files, err := baseFiles()
	if err != nil {
		return target.Output{}, err
	}
	files = append(files, generated...)
	files = append(files,
		render("package.json.tmpl", "package.json", struct{ Name string }{app.Name + "-" + part.Name}),
		render("prettierignore.tmpl", ".prettierignore", paths(generated)))
	slices.SortFunc(files, func(a, b target.File) int { return strings.Compare(a.Path, b.Path) })
	return target.Output{Files: files}, nil
}

// paths lists generated files for .prettierignore: philote owns their
// layout, and long lines would otherwise fight Prettier's. Brackets are
// escaped, since .prettierignore reads [id] as a glob.
func paths(files []target.File) []string {
	escape := strings.NewReplacer("[", `\[`, "]", `\]`)
	var out []string
	for _, f := range files {
		out = append(out, escape.Replace(f.Path))
	}
	slices.Sort(out)
	return out
}

// usedBackend finds the part a frontend's `uses` names. check guarantees
// it exists and is a backend.
func usedBackend(app *ir.App, part *ir.Part) *ir.Part {
	for _, p := range app.Parts {
		if p.Name == part.Uses && p.Kind == ir.Backend {
			return p
		}
	}
	panic("nextjs: " + part.Name + " uses " + part.Uses + ", which isn't a backend")
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
		panic("nextjs: template " + name + ": " + err.Error())
	}
	return target.File{Path: path, Data: b.Bytes()}
}
