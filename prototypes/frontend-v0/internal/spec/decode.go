package spec

import (
	"strconv"
	"strings"
)

// Stage 2: raw blocks -> typed AST. Each block kind has a fixed schema;
// unknown keys, missing keys and malformed values are errors here.

const maxCopyLen = 200 // cap for natural-language UI copy

var (
	appKeys       = []string{"target", "style"}
	pageKeys      = []string{"data", "layout", "empty", "not_found"}
	componentKeys = []string{"shows", "links_to"}
)

func ident(t Token) Ident { return Ident{t.Text, t.Pos} }

// vp parses one field's value tokens.
type vp struct {
	toks []Token
	i    int
	d    *Diags
	end  Pos
}

func newVP(f rawField, d *Diags) *vp { return &vp{toks: f.Value, d: d, end: f.End} }

func (v *vp) peek() Token {
	if v.i < len(v.toks) {
		return v.toks[v.i]
	}
	return Token{Kind: tNewline, Pos: v.end}
}

func (v *vp) next() Token {
	t := v.peek()
	if v.i < len(v.toks) {
		v.i++
	}
	return t
}

func (v *vp) atEnd() bool          { return v.i >= len(v.toks) }
func (v *vp) isWord(w string) bool { t := v.peek(); return t.Kind == tIdent && t.Text == w }

func (v *vp) fail(t Token, hint, format string, args ...any) {
	v.d.Errorf(t.Pos, format, args...)
	if hint != "" {
		v.d.Hint("%s", hint)
	}
	panic(bail{})
}

func (v *vp) expect(k tokKind, what string) Token {
	if t := v.peek(); t.Kind != k {
		v.fail(t, "", "expected %s, found %s", what, t.describe())
	}
	return v.next()
}

func (v *vp) expectWord(w string) Token {
	if t := v.peek(); t.Kind != tIdent || t.Text != w {
		v.fail(t, "", "expected '%s', found %s", w, t.describe())
	}
	return v.next()
}

func (v *vp) finish(hint, after string) {
	if !v.atEnd() {
		t := v.peek()
		v.fail(t, hint, "unexpected %s after %s", t.describe(), after)
	}
}

func try(fn func()) (ok bool) {
	ok = true
	defer recoverBail(&ok)
	fn()
	return
}

func decode(path, src string, blocks []*rawBlock, d *Diags) *File {
	f := &File{Path: path, Src: src}
	for _, b := range blocks {
		switch b.Kind {
		case "app":
			a := decodeApp(b, d)
			if f.App != nil {
				d.Errorf(b.Kw.Pos, "only one app block is allowed")
				d.Hint("the first one is on line %d", f.App.Pos.Line)
				continue
			}
			f.App = a
			f.Decls = append(f.Decls, a)
		case "entity":
			e := decodeEntity(b, d)
			f.Entities = append(f.Entities, e)
			f.Decls = append(f.Decls, e)
		case "page":
			pg := decodePage(b, d)
			f.Pages = append(f.Pages, pg)
			f.Decls = append(f.Decls, pg)
		case "component":
			c := decodeComponent(b, d)
			f.Components = append(f.Components, c)
			f.Decls = append(f.Decls, c)
		}
	}
	if f.App == nil {
		d.Errorf(Pos{1, 1}, "missing app block")
		d.Hint("start the file with `app <name> { target: nextjs  style: tokens/minimal }`")
	}
	return f
}

func unknownField(d *Diags, key Token, kind string, keys []string) {
	d.Errorf(key.Pos, "unknown field '%s' in %s", key.Text, kind)
	d.hintNames(key.Text, keys, kind+" fields")
}

func requireKeys(d *Diags, b *rawBlock, kind string, keys ...string) {
	have := map[string]bool{}
	for _, f := range b.Fields {
		have[f.Key.Text] = true
	}
	for _, k := range keys {
		if !have[k] {
			d.Errorf(b.Name.Pos, "%s is missing '%s'", kind, k)
		}
	}
}

func decodeApp(b *rawBlock, d *Diags) *App {
	a := &App{Pos: b.Kw.Pos, Name: ident(b.Name)}
	for _, f := range b.Fields {
		switch f.Key.Text {
		case "target":
			try(func() {
				v := newVP(f, d)
				t := v.expect(tIdent, "a target name")
				v.finish("", "the target")
				a.Target = &Ident{t.Text, t.Pos}
			})
		case "style":
			try(func() {
				v := newVP(f, d)
				first := v.expect(tIdent, "a style path like tokens/minimal")
				parts := []string{first.Text}
				for v.peek().Kind == tSlash {
					v.next()
					parts = append(parts, v.expect(tIdent, "a path segment").Text)
				}
				v.finish("", "the style path")
				a.Style = &Str{first.Pos, strings.Join(parts, "/")}
			})
		default:
			unknownField(d, f.Key, "app", appKeys)
		}
	}
	requireKeys(d, b, "app "+b.Name.Text, appKeys...)
	return a
}

func decodeEntity(b *rawBlock, d *Diags) *Entity {
	e := &Entity{Pos: b.Kw.Pos, Name: ident(b.Name)}
	for _, f := range b.Fields {
		try(func() { e.Fields = append(e.Fields, decodeEntityField(f, d)) })
	}
	return e
}

func decodeEntityField(f rawField, d *Diags) *Field {
	v := newVP(f, d)
	fd := &Field{Name: ident(f.Key)}
	t := v.peek()
	switch t.Kind {
	case tLBrack:
		v.next()
		el := v.expect(tIdent, "a type name")
		v.expect(tRBrack, "']'")
		fd.Type = TypeRef{t.Pos, el.Text, true}
	case tIdent:
		v.next()
		fd.Type = TypeRef{t.Pos, t.Text, false}
		if t.Text == "slug" && v.isWord("from") {
			v.next()
			src := v.expect(tIdent, "the field the slug is made from")
			fd.SlugFrom = &Ident{src.Text, src.Pos}
		}
		if v.peek().Kind == tLParen {
			v.next()
			v.expectWord("max")
			n := v.expect(tInt, "a number")
			v.expect(tRParen, "')'")
			fd.Max = &Int{n.Pos, atoi(n.Text)}
		}
	default:
		v.fail(t, "", "expected a type for '%s', found %s", f.Key.Text, t.describe())
	}
	for !v.atEnd() {
		m := v.next()
		switch {
		case m.Kind == tIdent && m.Text == "unique":
			fd.Unique = true
		case m.Kind == tIdent && m.Text == "optional":
			fd.Optional = true
		default:
			v.fail(m, "modifiers are: unique, optional", "unexpected %s in the type of '%s'", m.describe(), f.Key.Text)
		}
	}
	return fd
}

func decodePage(b *rawBlock, d *Diags) *Page {
	pg := &Page{Pos: b.Kw.Pos, Route: parseRoute(b.Name, d)}
	for _, f := range b.Fields {
		switch f.Key.Text {
		case "data":
			try(func() { pg.Data = decodeQuery(f, d) })
		case "layout":
			try(func() {
				v := newVP(f, d)
				l := &Layout{Pos: v.peek().Pos}
				if v.isWord("list") {
					v.next()
					v.expectWord("of")
					l.List = true
				}
				c := v.expect(tIdent, "a component name")
				l.Component = ident(c)
				v.finish("", "the layout component")
				pg.Layout = l
			})
		case "empty":
			pg.Empty = decodeCopy(f, d)
		case "not_found":
			try(func() {
				v := newVP(f, d)
				n := v.expect(tInt, "a status code like 404")
				v.finish("", "the status code")
				pg.NotFound = &Int{n.Pos, atoi(n.Text)}
			})
		default:
			unknownField(d, f.Key, "page", pageKeys)
		}
	}
	requireKeys(d, b, "page "+strconv.Quote(b.Name.Text), "data", "layout")
	return pg
}

func decodeCopy(f rawField, d *Diags) *Str {
	var out *Str
	try(func() {
		v := newVP(f, d)
		s := v.expect(tString, "a quoted string")
		v.finish("", "the string")
		if len(s.Text) > maxCopyLen {
			v.fail(s, "", "'%s' is %d characters; UI copy is capped at %d", f.Key.Text, len(s.Text), maxCopyLen)
		}
		out = &Str{s.Pos, s.Text}
	})
	return out
}

func decodeQuery(f rawField, d *Diags) *Query {
	v := newVP(f, d)
	e := v.expect(tIdent, "an entity name")
	q := &Query{Pos: e.Pos, Entity: ident(e)}
	if v.isWord("by") {
		v.next()
		k := v.expect(tIdent, "a field name")
		q.By = &Ident{k.Text, k.Pos}
		v.finish("a lookup by one field can't be combined with where/order by", "'by "+k.Text+"'")
		return q
	}
	if v.isWord("where") {
		v.next()
		for {
			fld := v.expect(tIdent, "a field name")
			op := v.expect(tOp, "a comparison (==, !=, <, <=, >, >=)")
			q.Where = append(q.Where, &Cond{Field: ident(fld), Op: op.Text, OpPos: op.Pos, Value: v.parseExpr()})
			if !v.isWord("and") {
				break
			}
			v.next()
		}
	}
	if v.isWord("order") {
		v.next()
		v.expectWord("by")
		fld := v.expect(tIdent, "a field name")
		o := &Order{Field: ident(fld)}
		if v.isWord("desc") {
			v.next()
			o.Desc = true
		} else if v.isWord("asc") {
			v.next()
		}
		q.Order = o
	}
	if v.isWord("limit") {
		v.next()
		n := v.expect(tInt, "a number")
		q.Limit = &Int{n.Pos, atoi(n.Text)}
	}
	v.finish("a query is: Entity [where ...] [order by field [asc|desc]] [limit n], or Entity by field", "the query")
	return q
}

func (v *vp) parseExpr() Expr {
	t := v.peek()
	switch t.Kind {
	case tInt:
		v.next()
		return &IntLit{t.Pos, atoi(t.Text)}
	case tString:
		v.next()
		return &StrLit{t.Pos, t.Text}
	case tIdent:
		v.next()
		if v.peek().Kind == tLParen {
			v.next()
			c := &CallExpr{Fn: ident(t)}
			if v.peek().Kind != tRParen {
				for {
					c.Args = append(c.Args, v.parseExpr())
					if v.peek().Kind != tComma {
						break
					}
					v.next()
				}
			}
			v.expect(tRParen, "')' or ','")
			return c
		}
		p := &PathExpr{Parts: []Ident{ident(t)}}
		for v.peek().Kind == tDot {
			v.next()
			p.Parts = append(p.Parts, ident(v.expect(tIdent, "a field name after '.'")))
		}
		return p
	}
	v.fail(t, "", "expected a value, found %s", t.describe())
	return nil
}

func decodeComponent(b *rawBlock, d *Diags) *Component {
	c := &Component{Pos: b.Kw.Pos, Name: ident(b.Name), Params: b.Params, Native: b.Native, From: b.From}
	for _, f := range b.Fields {
		switch f.Key.Text {
		case "shows":
			try(func() {
				v := newVP(f, d)
				for {
					s := &Show{Expr: v.parseExpr()}
					if v.isWord("as") {
						v.next()
						r := v.expect(tIdent, "a renderer name (markdown, chips)")
						s.As = &Ident{r.Text, r.Pos}
					}
					c.Shows = append(c.Shows, s)
					if v.peek().Kind != tComma {
						break
					}
					v.next()
				}
				v.finish("separate items with ','", "a shown item")
			})
		case "links_to":
			try(func() {
				v := newVP(f, d)
				s := v.expect(tString, `a quoted link like "/blog/[post.slug]"`)
				v.finish("", "the link")
				c.LinksTo = parseTemplate(s, d)
			})
		default:
			unknownField(d, f.Key, "component", componentKeys)
		}
	}
	if !b.Native {
		requireKeys(d, b, "component "+b.Name.Text, "shows")
	}
	return c
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func isIdentString(s string) bool {
	if s == "" || !isLetter(s[0]) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isLetter(s[i]) && !isDigit(s[i]) {
			return false
		}
	}
	return true
}

func isLiteralSeg(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z') && !isDigit(c) && c != '-' && c != '_' {
			return false
		}
	}
	return s != ""
}

// splitRoute yields each segment with its column in the source.
// tok.Pos points at the opening quote, so the raw string starts one column later.
func splitRoute(tok Token, d *Diags) (segs []string, cols []int, ok bool) {
	raw := tok.Text
	if !strings.HasPrefix(raw, "/") {
		d.Errorf(tok.Pos, "a route must start with '/'")
		return nil, nil, false
	}
	if raw == "/" {
		return nil, nil, true
	}
	off := 1
	for _, s := range strings.Split(raw[1:], "/") {
		segs = append(segs, s)
		cols = append(cols, tok.Pos.Col+1+off)
		off += len(s) + 1
	}
	return segs, cols, true
}

func parseRoute(tok Token, d *Diags) Route {
	r := Route{Pos: tok.Pos, Raw: tok.Text}
	segs, cols, ok := splitRoute(tok, d)
	if !ok {
		return r
	}
	for i, s := range segs {
		p := Pos{tok.Pos.Line, cols[i]}
		switch {
		case s == "":
			d.Errorf(p, "empty segment in route %q", tok.Text)
		case strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") && isIdentString(s[1:len(s)-1]):
			r.Segs = append(r.Segs, Seg{Pos: p, Param: s[1 : len(s)-1]})
		case isLiteralSeg(s):
			r.Segs = append(r.Segs, Seg{Pos: p, Lit: s})
		case strings.ContainsAny(s, "[]"):
			d.Errorf(p, "a route parameter must be a whole segment, like [slug]")
		default:
			d.Errorf(p, "route segment %q may only use a-z, 0-9, '-' and '_'", s)
		}
	}
	return r
}

func parseTemplate(tok Token, d *Diags) *Template {
	t := &Template{Pos: tok.Pos, Raw: tok.Text}
	segs, cols, ok := splitRoute(tok, d)
	if !ok {
		return t
	}
	for i, s := range segs {
		p := Pos{tok.Pos.Line, cols[i]}
		switch {
		case s == "":
			d.Errorf(p, "empty segment in link %q", tok.Text)
		case strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"):
			path := &PathExpr{}
			col := p.Col + 1
			for _, part := range strings.Split(s[1:len(s)-1], ".") {
				if !isIdentString(part) {
					d.Errorf(Pos{p.Line, col}, "expected a field path like [post.slug] in link %q", tok.Text)
					return t
				}
				path.Parts = append(path.Parts, Ident{part, Pos{p.Line, col}})
				col += len(part) + 1
			}
			t.Segs = append(t.Segs, TSeg{Pos: p, Path: path})
		case isLiteralSeg(s):
			t.Segs = append(t.Segs, TSeg{Pos: p, Lit: s})
		default:
			d.Errorf(p, "link segment %q may only use a-z, 0-9, '-', '_' or a [field.path]", s)
		}
	}
	return t
}
