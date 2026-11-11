package nextjs

import (
	"encoding/json"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// pageView is everything a page template prints. Title, Empty and NotFound
// are JavaScript string literals.
type pageView struct {
	Route, Title, Client, Read, Component, ComponentFile, Param string
	Empty, NotFound                                             string
}

type navView struct{ Title, Href string }

// pages renders each page, plus a not-found file for pages that load by id,
// and the nav that links to the list pages.
func pages(part *ir.Part) []target.File {
	var files []target.File
	var nav []navView
	for _, pg := range part.Pages {
		comp := component(part, pg.Component)
		v := pageView{
			Route:         pg.Route,
			Title:         jsString(title(pg)),
			Client:        camel(pg.Resource),
			Read:          "read" + comp.Name,
			Component:     comp.Name,
			ComponentFile: kebab(comp.Name),
			Param:         comp.Param,
			Empty:         jsString(pg.Empty),
			NotFound:      jsString(pg.NotFound),
		}
		dir := "src/app" + pg.Route + "/"
		if pg.ByID {
			files = append(files,
				render("record-page.tsx.tmpl", dir+"page.tsx", v),
				render("not-found.tsx.tmpl", dir+"not-found.tsx", v))
			continue
		}
		files = append(files, render("list-page.tsx.tmpl", dir+"page.tsx", v))
		nav = append(nav, navView{v.Title, jsString(pg.Route)})
	}
	return append(files, render("nav.ts.tmpl", "src/lib/nav.ts", nav))
}

// title is a list page's last segment, like "Horses" for /horses, or the
// entity's name for a page by id.
func title(pg *ir.Page) string {
	if pg.ByID {
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

// label turns matched_at into "Matched at".
func label(name string) string {
	words := strings.ReplaceAll(name, "_", " ")
	return strings.ToUpper(words[:1]) + words[1:]
}
