package syntax

// The statements that only go inside a frontend: what it calls, how it
// looks, its nav and its viewer.

func (p *parser) uses() Item {
	u := &Uses{Word: p.ident("uses")}
	p.expect(tColon, "':' and a backend part, like `uses: api`")
	u.Part = p.ident("a backend part name")
	p.endStatement()
	return u
}

func (p *parser) look() Item {
	l := &Look{Word: p.ident("look")}
	p.expect(tColon, "':' and how the app should look, in one sentence")
	l.Value = p.str("how the app should look, in one sentence")
	p.endStatement()
	return l
}

func (p *parser) theme() Item {
	t := &Theme{Word: p.ident("theme")}
	p.expect(tColon, "':' and a theme name, like `theme: warm`")
	t.Name = p.ident("a theme name: plain, warm, editorial or slate")
	p.endStatement()
	return t
}

func (p *parser) nav() Item {
	n := &Nav{Word: p.ident("nav")}
	p.expect(tColon, "':', where the nav sits and its pages, like `nav: bottom, \"/horses\" as \"Horses\"`")
	n.Place = p.ident("where the nav sits: top, bottom or side")
	for p.peek().kind == tComma {
		p.next()
		p.skipNewlines()
		it := NavItem{Route: p.str(`a page's route, like "/horses"`)}
		if p.isWord("as") {
			p.next()
			it.Label = p.str("the item's text, in quotes")
		}
		n.Items = append(n.Items, it)
	}
	p.endStatement()
	return n
}

func (p *parser) viewer() Item {
	v := &Viewer{Word: p.ident("viewer")}
	p.expect(tColon, "':' and the entity a visitor acts as, like `viewer: Horse`")
	v.Entity = p.ident("an entity, like Horse")
	p.endStatement()
	return v
}
