package check

import (
	"slices"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

var themes = []string{"plain", "warm", "editorial", "slate"}

// frontendSrc is a frontend's body, kept until every part is known: its
// pages read from a backend that may be declared after it.
type frontendSrc struct {
	part       *ir.Part
	uses       syntax.Ident
	pages      []*syntax.Page
	components []*syntax.Component
	nav        *syntax.Nav
	viewer     *syntax.Viewer
}

func (c *checker) frontend(part *ir.Part, p *syntax.Part) {
	src := &frontendSrc{part: part}
	var uses *syntax.Uses
	for _, it := range p.Body {
		switch x := it.(type) {
		case *syntax.Uses:
			if uses != nil {
				c.d.Errorf(x.Word.Pos, "%s already uses %s", part.Name, uses.Part.Name)
				continue
			}
			uses = x
		case *syntax.Theme:
			c.theme(part, x)
		case *syntax.Nav:
			if src.nav != nil {
				c.d.Errorf(x.Word.Pos, "%s already has a nav", part.Name)
				continue
			}
			src.nav = x
		case *syntax.Viewer:
			if src.viewer != nil {
				c.d.Errorf(x.Word.Pos, "%s already has a viewer", part.Name)
				continue
			}
			src.viewer = x
		case *syntax.Page:
			src.pages = append(src.pages, x)
		case *syntax.Component:
			src.components = append(src.components, x)
		default:
			c.d.Errorf(it.Start(), "a frontend takes uses, theme, nav, viewer, page and component")
		}
	}
	if uses == nil {
		c.d.Errorf(p.Name.Pos, "%s doesn't say which backend it calls", part.Name)
		c.d.Hint("add `uses: <backend>` inside its { }")
		return
	}
	src.uses = uses.Part
	c.frontends = append(c.frontends, src)
}

func (c *checker) resolveFrontends() {
	backends := map[string]*ir.Part{}
	var names []string
	for _, p := range c.app.Parts {
		if p.Kind == ir.Backend {
			backends[p.Name] = p
			names = append(names, p.Name)
		}
	}
	for _, f := range c.frontends {
		b := backends[f.uses.Name]
		if b == nil {
			c.d.Errorf(f.uses.Pos, "there is no backend part named %s", f.uses.Name)
			c.d.HintNames(f.uses.Name, names, "backend parts")
			continue
		}
		f.part.Uses = b.Name
		if f.viewer != nil {
			c.viewer(f.part, b, f.viewer)
		}
		c.views(f, b)
		c.picker(f.part)
		if f.nav != nil {
			c.nav(f.part, f.nav)
		}
	}
}

func (c *checker) theme(part *ir.Part, t *syntax.Theme) {
	switch {
	case part.Theme != "":
		c.d.Errorf(t.Word.Pos, "%s already has a theme", part.Name)
	case !slices.Contains(themes, t.Name.Name):
		c.d.Errorf(t.Name.Pos, "unknown theme '%s'", t.Name.Name)
		c.d.HintNames(t.Name.Name, themes, "themes")
	default:
		part.Theme = t.Name.Name
	}
}
