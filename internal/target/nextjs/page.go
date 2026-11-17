package nextjs

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// pageView is everything a page template prints. Title, Empty and
// NotFound are JavaScript string literals.
type pageView struct {
	Route, Title, NotFound string
	Client                 string // the client call that loads the page's data; empty for none
	ByID                   bool
	Viewer                 bool   // the page shows the viewer
	Async                  bool   // the page loads data or the viewer, so its sections stream in
	Reads                  string // imports from @/lib/reads
	Components             []componentImport
	Prep                   []string // statements that read each section's props
	Sections               []string // each section's JSX
}

type componentImport struct{ Name, File string }

// pages renders each page, plus a not-found file for pages that load by id.
// It also returns the default nav's items: the pages without params, then
// the forms that make a new record.
func pages(app *ir.App, backend, part *ir.Part) ([]target.File, []navItemView) {
	var files []target.File
	var nav, forms []navItemView
	for _, pg := range part.Pages {
		if pg.Form != "" {
			files = append(files, formFiles(app, backend, pg)...)
			if !pg.ByID {
				forms = append(forms, navItemView{jsString(pg.Route), jsString(formTitle(pg))})
			}
			continue
		}
		v := newPageView(app, part, pg)
		dir := "src/app" + strings.TrimSuffix(pg.Route, "/") + "/"
		files = append(files, render("page.tsx.tmpl", dir+"page.tsx", v))
		switch {
		case pg.ByID:
			files = append(files, render("not-found.tsx.tmpl", dir+"not-found.tsx", v))
		case pg.Route != "/" && !strings.Contains(pg.Route, "["):
			nav = append(nav, navItemView{jsString(pg.Route), v.Title})
		}
	}
	return files, append(nav, forms...)
}

func newPageView(app *ir.App, part *ir.Part, pg *ir.Page) pageView {
	v := pageView{Route: pg.Route, Title: jsString(title(app, pg)), NotFound: jsString(pg.NotFound), ByID: pg.ByID}
	if pg.Load != "" {
		// The page holds what it loads in a local named loaded and reaches
		// the client through the api namespace, so no spec name, like a load
		// named horses next to the horses client, can shadow generated code.
		v.Client = "api." + camel(pg.Resource) + ".list()"
		if pg.ByID {
			v.Client = "api." + camel(pg.Resource) + ".get(id)"
		}
	}
	reads, imported := map[string]bool{}, map[string]bool{}
	for i, sec := range pg.Sections {
		c := component(part, sec.Component)
		if !imported[c.Name] {
			v.Components = append(v.Components, componentImport{c.Name, kebab(c.Name)})
			imported[c.Name] = true
		}
		if !sec.Arg {
			v.Sections = append(v.Sections, "<"+c.Name+" />")
			continue
		}
		reads["read"+c.Name] = true
		props := fmt.Sprintf("props%d", i+1)
		v.Prep = append(v.Prep, sectionPrep(c, sec, props))
		v.Sections = append(v.Sections, sectionJSX(pg, c, sec, props))
	}
	v.Reads = strings.Join(sortedKeys(reads), ", ")
	v.Viewer = showsViewer(pg)
	v.Async = v.Client != "" || v.Viewer
	return v
}

func showsViewer(pg *ir.Page) bool {
	return slices.ContainsFunc(pg.Sections, func(s *ir.Section) bool { return s.Viewer })
}

// sectionPrep reads a section's props from what the page loaded.
func sectionPrep(c *ir.Component, sec *ir.Section, props string) string {
	if sec.Viewer {
		return fmt.Sprintf("const %s = await read%s(viewer);", props, c.Name)
	}
	if sec.List {
		return fmt.Sprintf("const %s = await Promise.all(loaded.map(read%s));", props, c.Name)
	}
	return fmt.Sprintf("const %s = await read%s(loaded);", props, c.Name)
}

// sectionJSX renders a section. A list that's empty shows the page's empty
// text in its place.
func sectionJSX(pg *ir.Page, c *ir.Component, sec *ir.Section, props string) string {
	empty := `<p className="mx-auto w-full max-w-5xl px-4 py-16 text-center text-muted-foreground sm:px-6">{` + jsString(pg.Empty) + `}</p>`
	switch {
	case sec.Viewer:
	case sec.List:
		return fmt.Sprintf(`{%[2]s.length === 0 ? (
                %[3]s
            ) : (
                <section className="mx-auto grid w-full max-w-5xl gap-6 px-4 py-8 sm:grid-cols-2 sm:px-6 lg:grid-cols-3">
                    {%[2]s.map((item) => (
                        <%[1]s key={item.%[4]s.id} {...item} />
                    ))}
                </section>
            )}`, c.Name, props, empty, c.Param)
	case !pg.ByID:
		return fmt.Sprintf(`{%[2]s.%[3]s.length === 0 ? %[4]s : <%[1]s {...%[2]s} />}`, c.Name, props, c.Param, empty)
	}
	return fmt.Sprintf("<%s {...%s} />", c.Name, props)
}

// pageTitle is what a page calls itself in its heading and the nav.
func pageTitle(app *ir.App, pg *ir.Page) string {
	if pg.Form != "" {
		return formTitle(pg)
	}
	return title(app, pg)
}

// title is a page's last segment, like "Horses" for /horses, the entity's
// name for a page by id, or the app's name for /.
func title(app *ir.App, pg *ir.Page) string {
	switch {
	case pg.Route == "/":
		return app.Name
	case pg.ByID:
		return pg.Entity
	}
	seg := pg.Route[strings.LastIndex(pg.Route, "/")+1:]
	words := strings.ReplaceAll(seg, "-", " ")
	return strings.ToUpper(words[:1]) + words[1:]
}

func component(part *ir.Part, name string) *ir.Component {
	for _, c := range part.Components {
		if c.Name == name {
			return c
		}
	}
	panic("nextjs: no component " + name)
}

// jsString quotes s as a JavaScript string literal. JSON's escapes are a
// subset of JavaScript's, unlike Go's.
func jsString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// kebab turns HorseCard into horse-card.
func kebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}
