package syntax

func (p *parser) page() Item {
	pg := &Page{Word: p.ident("page")}
	pg.Route = p.str(`a route like "/horses"`)
	p.keyed("page", []string{"load", "shows", "empty", "not_found", "form", "then"}, func(key Ident) {
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
		case "load":
			pg.Load = p.load()
		case "shows":
			pg.Shows = append(pg.Shows, p.section())
			for p.peek().kind == tComma {
				p.next()
				p.skipNewlines()
				pg.Shows = append(pg.Shows, p.section())
			}
		case "empty":
			pg.Empty = p.str("what to show when there are no records, in quotes")
		case "not_found":
			pg.NotFound = p.str("what to show when the record doesn't exist, in quotes")
		}
	})
	p.endStatement()
	return pg
}

// load is `<name> [from <resource>] [by <field>]` or `<name> from
// <endpoint>(<args>)`.
func (p *parser) load() *PageLoad {
	l := &PageLoad{Name: p.ident("a name for what the page loads, like horses")}
	if p.isWord("from") {
		p.next()
		from := p.ident("a resource or endpoint name")
		l.From = &from
		if p.peek().kind == tLParen {
			l.Args, l.Call = p.args(), true
			return l
		}
	}
	l.By = p.by()
	return l
}

// args is `(<input>[: <value>], ...)`, possibly empty.
func (p *parser) args() []Arg {
	p.expect(tLParen, "'('")
	var out []Arg
	for p.peek().kind != tRParen {
		a := Arg{Name: p.ident("an input name")}
		if p.peek().kind == tColon {
			p.next()
			v := p.ident("what fills it, like viewer")
			a.Value = &v
		}
		out = append(out, a)
		if p.peek().kind != tComma {
			break
		}
		p.next()
	}
	p.expect(tRParen, "',' or ')'")
	return out
}

// section is `[list of] <Component>[(<load>)]`.
func (p *parser) section() Section {
	s := Section{List: p.listOf()}
	s.Component = p.ident("a component name")
	if p.peek().kind == tLParen {
		p.next()
		arg := p.ident("what the page loads, by name")
		s.Arg = &arg
		p.expect(tRParen, "')'")
	}
	return s
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
	if p.peek().kind == tLParen {
		c.Params = p.componentParams()
	}
	if p.peek().kind == tColon {
		p.next()
		k := p.ident("a component kind: card, detail, table, list, hero, callout, picker or deck")
		c.Kind = &k
	}
	p.keyed("component", []string{"shows", "says", "links_to", "look", "actions", "picks", "calls"}, func(key Ident) {
		switch key.Name {
		case "shows":
			c.Shows = append(c.Shows, p.show())
			for p.peek().kind == tComma {
				p.next()
				c.Shows = append(c.Shows, p.show())
			}
		case "says":
			c.Says = append(c.Says, p.str("words the component shows, in quotes"))
			for p.peek().kind == tComma {
				p.next()
				p.skipNewlines()
				c.Says = append(c.Says, p.str("words the component shows, in quotes"))
			}
		case "links_to":
			c.LinksTo = p.str(`a route like "/horses/[horse.id]"`)
			if p.isWord("as") {
				p.next()
				c.LinkLabel = p.str("the button's text, in quotes")
			}
		case "look":
			c.Look = p.str("how it should look, in one sentence")
		case "picks":
			picks := p.ident("what tapping a record picks: viewer")
			c.Picks = &picks
			p.word("then")
			c.PicksThen = p.str(`the route to go to after picking, like "/discover"`)
		case "calls":
			c.Calls = &Call{Endpoint: p.ident("an endpoint, like swipe")}
			c.Calls.Args = p.args()
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

// componentParams is `(<name>: [list of] <Entity>, ...)`.
func (p *parser) componentParams() []*Field {
	p.expect(tLParen, "'('")
	var out []*Field
	for p.peek().kind != tRParen {
		f := &Field{Name: p.ident("a parameter name")}
		p.expect(tColon, "':' and a type after '"+f.Name.Name+"'")
		f.List = p.listOf()
		f.Type = p.ident("an entity, like Horse")
		out = append(out, f)
		if p.peek().kind != tComma {
			break
		}
		p.next()
	}
	p.expect(tRParen, "',' or ')'")
	return out
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

// action is `open "<route>"` or `delete`, then optional `then "<route>"`
// and `as "<label>"`.
func (p *parser) action() Action {
	a := Action{Verb: p.ident("an action: open or delete")}
	if a.Verb.Name == "open" {
		a.Route = p.str(`a route like "/horses/[horse.id]/edit"`)
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
