package spec

// Stage 1: tokens -> raw blocks. A block is a keyword, a name, optional
// params, and a list of `key: value...` fields. Field values stay as tokens
// here; decode.go interprets them against each block kind's schema.

type rawBlock struct {
	Kw     Token
	Kind   string // app, entity, page, component
	Native bool
	Name   Token
	Params []*Param
	From   *Str
	Fields []rawField
}

type rawField struct {
	Key   Token
	Value []Token
	End   Pos // just past the last value token
}

type bail struct{}

func recoverBail(ok *bool) {
	if r := recover(); r != nil {
		if _, isBail := r.(bail); !isBail {
			panic(r)
		}
		*ok = false
	}
}

var declKeywords = map[string]bool{"app": true, "entity": true, "page": true, "component": true, "native": true}

type parser struct {
	toks []Token
	i    int
	d    *Diags
}

func (p *parser) peek() Token { return p.toks[p.i] }

func (p *parser) next() Token {
	t := p.toks[p.i]
	if t.Kind != tEOF {
		p.i++
	}
	return t
}

func (p *parser) skipNewlines() {
	for p.peek().Kind == tNewline {
		p.next()
	}
}

func (p *parser) fail(t Token, hint, format string, args ...any) {
	p.d.Errorf(t.Pos, format, args...)
	if hint != "" {
		p.d.Hint("%s", hint)
	}
	panic(bail{})
}

func (p *parser) expect(k tokKind, what string) Token {
	if t := p.peek(); t.Kind != k {
		p.fail(t, "", "expected %s, found %s", what, t.describe())
	}
	return p.next()
}

func (p *parser) expectWord(w, context string) Token {
	if t := p.peek(); t.Kind != tIdent || t.Text != w {
		p.fail(t, "", "expected '%s' %s, found %s", w, context, t.describe())
	}
	return p.next()
}

func (p *parser) parseBlocks() []*rawBlock {
	var out []*rawBlock
	for {
		p.skipNewlines()
		if p.peek().Kind == tEOF {
			return out
		}
		if b := p.parseBlockSafe(); b != nil {
			out = append(out, b)
		}
	}
}

func (p *parser) parseBlockSafe() (b *rawBlock) {
	start := p.i
	ok := true
	func() {
		defer recoverBail(&ok)
		b = p.parseBlock()
	}()
	if !ok {
		p.sync(start)
		return nil
	}
	return b
}

// sync skips to the next declaration keyword at the start of a line.
func (p *parser) sync(start int) {
	if p.i == start {
		p.next()
	}
	for {
		t := p.peek()
		if t.Kind == tEOF || (t.Kind == tIdent && t.Pos.Col == 1 && declKeywords[t.Text]) {
			return
		}
		p.next()
	}
}

func (p *parser) parseBlock() *rawBlock {
	kw := p.peek()
	if kw.Kind != tIdent || !declKeywords[kw.Text] {
		p.fail(kw, "", "expected a declaration (app, entity, page, component), found %s", kw.describe())
	}
	p.next()
	b := &rawBlock{Kw: kw, Kind: kw.Text}
	if kw.Text == "native" {
		p.expectWord("component", "after 'native'")
		b.Kind, b.Native = "component", true
	}
	if b.Kind == "page" {
		b.Name = p.expect(tString, `a route string like "/blog"`)
	} else {
		b.Name = p.expect(tIdent, "a "+b.Kind+" name")
	}
	if b.Kind == "component" {
		b.Params = p.parseParams()
	}
	if b.Native {
		p.expectWord("from", "and a file path after a native component's parameters")
		s := p.expect(tString, "a file path string")
		b.From = &Str{s.Pos, s.Text}
		if t := p.peek(); t.Kind != tNewline && t.Kind != tEOF {
			p.fail(t, "", "a native component has no body; unexpected %s", t.describe())
		}
		return b
	}

	p.expect(tLBrace, "'{'")
	seen := map[string]Token{}
	for {
		p.skipNewlines()
		t := p.peek()
		if t.Kind == tRBrace {
			p.next()
			return b
		}
		if t.Kind == tEOF {
			p.fail(kw, "", "this %s block is never closed; expected '}'", b.Kind)
		}
		key := p.expect(tIdent, "a field name")
		if prev, dup := seen[key.Text]; dup {
			p.d.Errorf(key.Pos, "'%s' is set twice", key.Text)
			p.d.Hint("first set on line %d", prev.Pos.Line)
		}
		seen[key.Text] = key
		p.expect(tColon, "':' after '"+key.Text+"'")
		f := rawField{Key: key}
		depth := 0
		for {
			t := p.peek()
			if t.Kind == tEOF || (depth == 0 && (t.Kind == tNewline || t.Kind == tRBrace)) {
				break
			}
			switch t.Kind {
			case tLParen, tLBrack:
				depth++
			case tRParen, tRBrack:
				depth = max(0, depth-1)
			}
			f.Value = append(f.Value, p.next())
		}
		if len(f.Value) == 0 {
			p.d.Errorf(key.Pos, "'%s' has no value", key.Text)
			continue
		}
		last := f.Value[len(f.Value)-1]
		f.End = Pos{last.Pos.Line, last.Pos.Col + last.width()}
		b.Fields = append(b.Fields, f)
	}
}

func (p *parser) parseParams() []*Param {
	p.expect(tLParen, "'(' and a parameter list")
	var ps []*Param
	if p.peek().Kind == tRParen {
		p.next()
		return ps
	}
	for {
		name := p.expect(tIdent, "a parameter name")
		p.expect(tColon, "':' and a type after '"+name.Text+"'")
		var tr TypeRef
		if t := p.peek(); t.Kind == tLBrack {
			p.next()
			el := p.expect(tIdent, "a type name")
			p.expect(tRBrack, "']'")
			tr = TypeRef{Pos: t.Pos, Name: el.Text, List: true}
		} else {
			el := p.expect(tIdent, "a type name")
			tr = TypeRef{Pos: el.Pos, Name: el.Text}
		}
		ps = append(ps, &Param{Name: Ident{name.Text, name.Pos}, Type: tr})
		if p.peek().Kind == tComma {
			p.next()
			continue
		}
		p.expect(tRParen, "',' or ')'")
		return ps
	}
}
