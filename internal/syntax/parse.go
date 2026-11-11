package syntax

import (
	"strconv"

	"github.com/3nd3r1/philote/internal/diag"
)

// bail unwinds out of a statement after its error has been reported.
type bail struct{}

type parser struct {
	toks []token
	i    int
	d    *diag.List
}

// Parse parses one file. Errors go to d; the result holds every statement
// that parsed cleanly.
func Parse(name, src string, d *diag.List) *File {
	p := &parser{toks: lex(name, src, d), d: d}
	return &File{Name: name, Items: p.items(false)}
}

func (p *parser) peek() token { return p.toks[p.i] }

func (p *parser) peekAt(n int) token { return p.toks[min(p.i+n, len(p.toks)-1)] }

func (p *parser) next() token {
	t := p.toks[p.i]
	if t.kind != tEOF {
		p.i++
	}
	return t
}

func (p *parser) skipNewlines() {
	for p.peek().kind == tNewline {
		p.next()
	}
}

func (p *parser) fail(t token, hint, format string, args ...any) {
	p.d.Errorf(t.pos, format, args...)
	if hint != "" {
		p.d.Hint("%s", hint)
	}
	panic(bail{})
}

func (p *parser) expect(k tokKind, what string) token {
	if t := p.peek(); t.kind != k {
		p.fail(t, "", "expected %s, found %s", what, t.describe())
	}
	return p.next()
}

func (p *parser) ident(what string) Ident {
	t := p.expect(tIdent, what)
	return Ident{t.text, t.pos}
}

func (p *parser) word(w string) {
	if t := p.peek(); t.kind != tIdent || t.text != w {
		p.fail(t, "", "expected '%s', found %s", w, t.describe())
	}
	p.next()
}

func (p *parser) isWord(w string) bool {
	t := p.peek()
	return t.kind == tIdent && t.text == w
}

// endStatement requires the statement to end here: a newline, '}' or EOF.
func (p *parser) endStatement() {
	switch t := p.peek(); t.kind {
	case tNewline:
		p.next()
	case tRBrace, tEOF:
	default:
		p.fail(t, "", "unexpected %s; each statement goes on its own line", t.describe())
	}
}

// items parses statements until '}' (inBlock) or end of file.
func (p *parser) items(inBlock bool) []Item {
	var out []Item
	for {
		p.skipNewlines()
		switch t := p.peek(); t.kind {
		case tEOF:
			if inBlock {
				p.fail(t, "", "expected '}' before the end of the file")
			}
			return out
		case tRBrace:
			if inBlock {
				return out
			}
			p.d.Errorf(t.pos, "unexpected '}'")
			p.next()
		default:
			if it := p.statement(); it != nil {
				out = append(out, it)
			}
		}
	}
}

// statement parses one statement, recovering to the next line on error.
func (p *parser) statement() (it Item) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(bail); !ok {
				panic(r)
			}
			p.skipStatement()
			it = nil
		}
	}()
	return p.item()
}

// skipStatement skips to the end of the current line, including any block it opens.
func (p *parser) skipStatement() {
	depth := 0
	for {
		switch p.peek().kind {
		case tEOF:
			return
		case tNewline:
			if depth == 0 {
				return
			}
		case tLBrace:
			depth++
		case tRBrace:
			if depth == 0 {
				return
			}
			depth--
		}
		p.next()
	}
}

const statementHint = "statements are app, backend, frontend, enum, entity, resource, endpoint, step, uses, or a relation like `Order has many lines: OrderLine`"

func (p *parser) item() Item {
	t := p.peek()
	if t.kind != tIdent {
		p.fail(t, statementHint, "expected a statement, found %s", t.describe())
	}
	switch t.text {
	case "app":
		return p.app()
	case "backend", "frontend":
		return p.part()
	case "enum":
		return p.enum()
	case "entity":
		return p.entity()
	case "resource":
		return p.resource()
	case "endpoint":
		return p.endpoint()
	case "step":
		return p.step()
	case "uses":
		return p.uses()
	}
	if v := p.peekAt(1); v.kind == tIdent && (v.text == "has" || v.text == "belongs") {
		return p.relation()
	}
	p.fail(t, statementHint, "unknown statement '%s'", t.text)
	return nil
}

func (p *parser) block() []Item {
	p.expect(tLBrace, "'{'")
	items := p.items(true)
	p.expect(tRBrace, "'}'")
	return items
}

func (p *parser) app() Item {
	p.word("app")
	a := &App{Name: p.ident("an app name")}
	a.Items = p.block()
	p.endStatement()
	return a
}

func (p *parser) part() Item {
	pt := &Part{Kind: p.ident("backend or frontend")}
	pt.Name = p.ident("a name")
	p.expect(tColon, "':' and a target, like `backend api: hono`")
	pt.Target = p.ident("a target name")
	if p.peek().kind == tLBrace {
		pt.HasBody = true
		pt.Body = p.block()
	}
	p.endStatement()
	return pt
}

func (p *parser) uses() Item {
	u := &Uses{Word: p.ident("uses")}
	p.expect(tColon, "':' and a backend part, like `uses: api`")
	u.Part = p.ident("a backend part name")
	p.endStatement()
	return u
}

func (p *parser) enum() Item {
	p.word("enum")
	e := &Enum{Name: p.ident("an enum name")}
	p.expect(tLBrace, "'{'")
	for {
		p.skipNewlines()
		e.Values = append(e.Values, p.ident("an enum value"))
		p.skipNewlines()
		if p.peek().kind != tComma {
			break
		}
		p.next()
	}
	p.expect(tRBrace, "',' or '}'")
	p.endStatement()
	return e
}

func (p *parser) entity() Item {
	p.word("entity")
	e := &Entity{Name: p.ident("an entity name")}
	p.expect(tLBrace, "'{'")
	for {
		p.skipNewlines()
		if p.peek().kind == tRBrace {
			break
		}
		p.recoverLine(func() {
			if p.isWord("display") {
				e.Displays = append(e.Displays, p.display())
				return
			}
			e.Fields = append(e.Fields, p.field())
		})
	}
	p.expect(tRBrace, "'}'")
	p.endStatement()
	return e
}

func (p *parser) display() Display {
	d := Display{Word: p.ident("display")}
	p.expect(tColon, "':' and a field, like `display: name`")
	d.Field = p.ident("a field name")
	p.endStatement()
	return d
}

// recoverLine runs fn, and on a parse error skips to the next line so the
// rest of the block still gets parsed.
func (p *parser) recoverLine(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(bail); !ok {
				panic(r)
			}
			p.skipStatement()
		}
	}()
	fn()
}

func (p *parser) field() *Field {
	f := &Field{Name: p.ident("a field name")}
	p.expect(tColon, "':' and a type after '"+f.Name.Name+"'")
	f.Type = p.ident("a type")
	if p.peek().kind == tLParen {
		f.Limits = p.limits()
	}
	if p.isWord("from") {
		p.next()
		from := p.ident("the field to make the slug from")
		f.From = &from
	}
	for p.peek().kind == tIdent {
		f.Mods = append(f.Mods, p.ident("a modifier"))
	}
	p.endStatement()
	return f
}

func (p *parser) limits() []Limit {
	p.expect(tLParen, "'('")
	var out []Limit
	for {
		name := p.ident("min or max")
		n := p.expect(tInt, "a number")
		v, _ := strconv.Atoi(n.text)
		out = append(out, Limit{name, v})
		if p.peek().kind != tComma {
			break
		}
		p.next()
	}
	p.expect(tRParen, "',' or ')'")
	return out
}

func (p *parser) relation() Item {
	r := &Relation{From: p.ident("an entity name")}
	verb := p.next()
	switch verb.text {
	case "has":
		p.word("many")
		r.Verb = Ident{"has many", verb.pos}
	case "belongs":
		p.word("to")
		r.Verb = Ident{"belongs to", verb.pos}
	}
	r.Name = p.ident("the relation's name, like `lines: OrderLine`")
	p.expect(tColon, "':' and an entity after '"+r.Name.Name+"'")
	r.To = p.ident("an entity name")
	if p.peek().kind == tLParen {
		p.next()
		p.word("back")
		back := p.ident("the name of the reverse side")
		r.Back = &back
		p.expect(tRParen, "')'")
	}
	p.endStatement()
	return r
}

func (p *parser) resource() Item {
	p.word("resource")
	r := &Resource{Name: p.ident("a resource name")}
	p.expect(tColon, "':' and an entity, like `resource orders: Order`")
	r.Entity = p.ident("an entity name")
	if p.peek().kind == tLBrace {
		p.next()
		p.skipNewlines()
		if k := p.peek(); !p.isWord("only") {
			p.fail(k, "the only option is `only: list, get, create, update, delete`", "unknown resource option %s", k.describe())
		}
		p.next()
		p.expect(tColon, "':'")
		r.Only = append(r.Only, p.ident("an operation"))
		for p.peek().kind == tComma {
			p.next()
			r.Only = append(r.Only, p.ident("an operation"))
		}
		p.skipNewlines()
		p.expect(tRBrace, "'}'")
	}
	p.endStatement()
	return r
}
