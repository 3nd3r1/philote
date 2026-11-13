package syntax

import "slices"

// keyed parses a `{ key: value ... }` block where each key appears at most
// once and is one of keys. value parses what follows the colon.
func (p *parser) keyed(what string, keys []string, value func(key Ident)) {
	p.expect(tLBrace, "'{'")
	seen := map[string]bool{}
	for {
		p.skipNewlines()
		if p.peek().kind == tRBrace {
			break
		}
		p.recoverLine(func() {
			key := p.ident("a key")
			switch {
			case !slices.Contains(keys, key.Name):
				p.d.Errorf(key.Pos, "unknown key '%s' in %s", key.Name, what)
				p.d.HintNames(key.Name, keys, what+" keys")
				panic(bail{})
			case seen[key.Name]:
				p.d.Errorf(key.Pos, "'%s' is set twice", key.Name)
				panic(bail{})
			}
			seen[key.Name] = true
			p.expect(tColon, "':' after '"+key.Name+"'")
			value(key)
			p.endStatement()
		})
	}
	p.expect(tRBrace, "'}'")
}

func (p *parser) str(what string) *Str {
	t := p.expect(tString, what)
	return &Str{t.text, t.pos}
}

func (p *parser) identList(what string) []Ident {
	out := []Ident{p.ident(what)}
	for p.peek().kind == tComma {
		p.next()
		out = append(out, p.ident(what))
	}
	return out
}

// param is `<name>: <Type>(limits)`.
func (p *parser) param() *Field {
	f := &Field{Name: p.ident("a name")}
	p.expect(tColon, "':' and a type after '"+f.Name.Name+"'")
	f.Type = p.ident("a type")
	if p.peek().kind == tLParen {
		f.Limits = p.limits()
	}
	return f
}

// params is `(<param>, ...)`, possibly empty.
func (p *parser) params() []*Field {
	p.expect(tLParen, "'('")
	var out []*Field
	for p.peek().kind != tRParen {
		out = append(out, p.param())
		if p.peek().kind != tComma {
			break
		}
		p.next()
	}
	p.expect(tRParen, "',' or ')'")
	return out
}

// record is `{ <param>, <param> }`, with commas or newlines between.
func (p *parser) record() []*Field {
	p.expect(tLBrace, "'{'")
	var out []*Field
	for {
		p.skipNewlines()
		if p.peek().kind == tRBrace {
			break
		}
		out = append(out, p.param())
		p.skipNewlines()
		if p.peek().kind == tComma {
			p.next()
		}
	}
	p.expect(tRBrace, "'}'")
	return out
}

func (p *parser) endpoint() Item {
	p.word("endpoint")
	e := &Endpoint{Name: p.ident("an endpoint name")}
	p.keyed("endpoint", []string{"route", "input", "output", "errors", "steps"}, func(key Ident) {
		switch key.Name {
		case "route":
			m := p.ident("a method like POST")
			e.Method = &m
			e.Path = p.str(`a path like "/swipe"`)
		case "input":
			e.Input = p.record()
		case "output":
			o := p.ident("an entity")
			e.Output = &o
		case "errors":
			e.Errors = p.identList("an error name")
		case "steps":
			e.Steps = p.identList("a step name")
		}
	})
	p.endStatement()
	return e
}

func (p *parser) step() Item {
	p.word("step")
	s := &Step{Name: p.ident("a step name")}
	s.Params = p.params()
	if p.peek().kind == tArrow {
		p.next()
		s.Result = p.param()
	}
	p.keyed("step", []string{"fails", "reads", "writes", "intent"}, func(key Ident) {
		switch key.Name {
		case "fails":
			s.Fails = p.identList("an error name")
		case "reads":
			s.Reads = p.identList("an entity")
		case "writes":
			s.Writes = p.identList("an entity")
		case "intent":
			s.Intent = p.str(`a short description in quotes`)
		}
	})
	p.endStatement()
	return s
}
