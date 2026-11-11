package check

import (
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// frontendSrc is a frontend's body, kept until every part is known: its
// pages read from a backend that may be declared after it.
type frontendSrc struct {
	part       *ir.Part
	uses       syntax.Ident
	pages      []*syntax.Page
	components []*syntax.Component
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
		case *syntax.Page:
			src.pages = append(src.pages, x)
		case *syntax.Component:
			src.components = append(src.components, x)
		default:
			c.d.Errorf(it.Start(), "a frontend takes uses, page and component")
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
		c.views(f, b)
	}
}
