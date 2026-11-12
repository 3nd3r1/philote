package syntax

func (p *parser) page() Item {
	pg := &Page{Word: p.ident("page")}
	pg.Route = p.str(`a route like "/horses"`)
	p.keyed("page", []string{"data", "layout", "empty", "not_found", "form", "then"}, func(key Ident) {
		switch key.Name {
		case "form":
			f := p.ident("an endpoint, like horses.create")
			pg.Form = &f
			if p.peek().kind == tDot {
				p.next()
				op := p.ident("an operation, like create")
				pg.FormOp = &op
			}
			pg.By = p.by()
		case "then":
			pg.Then = p.str(`the route to go to after saving, like "/horses/[id]"`)
		case "data":
			d := p.ident("a resource name")
			pg.Data = &d
			pg.By = p.by()
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

// by is an optional `by <field>`.
func (p *parser) by() *Ident {
	if !p.isWord("by") {
		return nil
	}
	p.next()
	by := p.ident("a field, like `by id`")
	return &by
}

func (p *parser) component() Item {
	p.word("component")
	c := &Component{Name: p.ident("a component name")}
	c.Params = p.params()
	p.keyed("component", []string{"shows", "links_to", "look", "actions"}, func(key Ident) {
		switch key.Name {
		case "shows":
			c.Shows = append(c.Shows, p.show())
			for p.peek().kind == tComma {
				p.next()
				c.Shows = append(c.Shows, p.show())
			}
		case "links_to":
			c.LinksTo = p.str(`a route like "/horses/[horse.id]"`)
		case "look":
			c.Look = p.str("how it should look, in one sentence")
		case "actions":
			c.Actions = append(c.Actions, p.action())
			for p.peek().kind == tComma {
				p.next()
				p.skipNewlines()
				c.Actions = append(c.Actions, p.action())
			}
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

// action is `open "<route>"`, `delete` or `<endpoint>(<binds>)`, then
// optional `then "<route>"` and `as "<label>"`.
func (p *parser) action() Action {
	a := Action{Verb: p.ident("an action: open, delete, or an endpoint's name")}
	switch {
	case a.Verb.Name == "open":
		a.Route = p.str(`a route like "/horses/[horse.id]/edit"`)
	case p.peek().kind == tLParen:
		p.next()
		for p.peek().kind != tRParen {
			if len(a.Binds) > 0 {
				p.expect(tComma, "',' or ')'")
			}
			a.Binds = append(a.Binds, p.bind())
		}
		p.next()
	}
	if p.isWord("then") {
		p.next()
		a.Then = p.str(`the route to go to afterwards, like "/horses"`)
	}
	if p.isWord("as") {
		p.next()
		a.Label = p.str("the button's text, in quotes")
	}
	return a
}

// bind is `<input>: <value>`.
func (p *parser) bind() Bind {
	b := Bind{Name: p.ident("an input name")}
	p.expect(tColon, "':' after '"+b.Name.Name+"'")
	switch t := p.peek(); t.kind {
	case tString:
		b.Str = p.str("a string")
	case tInt:
		p.next()
		b.Int = &Ident{t.text, t.pos}
	default:
		b.Path = []Ident{p.ident("a value: a path like book.author, true, false, a number or a string")}
		for p.peek().kind == tDot {
			p.next()
			b.Path = append(b.Path, p.ident("a name after '.'"))
		}
	}
	return b
}
