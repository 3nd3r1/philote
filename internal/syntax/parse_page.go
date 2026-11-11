package syntax

func (p *parser) page() Item {
	pg := &Page{Word: p.ident("page")}
	pg.Route = p.str(`a route like "/horses"`)
	p.keyed("page", []string{"data", "layout", "empty", "not_found"}, func(key Ident) {
		switch key.Name {
		case "data":
			d := p.ident("a resource name")
			pg.Data = &d
			if p.isWord("by") {
				p.next()
				by := p.ident("a field, like `by id`")
				pg.By = &by
			}
		case "layout":
			if p.isWord("list") {
				p.next()
				p.word("of")
				pg.List = true
			}
			l := p.ident("a component name")
			pg.Layout = &l
		case "empty":
			pg.Empty = p.str("what to show when there are no records, in quotes")
		case "not_found":
			pg.NotFound = p.str("what to show when the record doesn't exist, in quotes")
		}
	})
	p.endStatement()
	return pg
}

func (p *parser) component() Item {
	p.word("component")
	c := &Component{Name: p.ident("a component name")}
	c.Params = p.params()
	p.keyed("component", []string{"shows", "links_to"}, func(key Ident) {
		switch key.Name {
		case "shows":
			c.Shows = append(c.Shows, p.show())
			for p.peek().kind == tComma {
				p.next()
				c.Shows = append(c.Shows, p.show())
			}
		case "links_to":
			c.LinksTo = p.str(`a route like "/horses/[horse.id]"`)
		}
	})
	p.endStatement()
	return c
}

// show is `<name>.<name>... [as <format>]`.
func (p *parser) show() Show {
	s := Show{Path: []Ident{p.ident("a path like horse.name")}}
	for p.peek().kind == tDot {
		p.next()
		s.Path = append(s.Path, p.ident("a name after '.'"))
	}
	if p.isWord("as") {
		p.next()
		f := p.ident("a format, like chips")
		s.Format = &f
	}
	return s
}
