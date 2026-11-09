package spec

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Stage 3: name resolution and type checking. Everything here happens before
// any model call. The checker also records which nodes each node actually
// depends on, which the build graph uses for cache keys.

type TKind int

const (
	TString TKind = iota
	TSlug
	TMarkdown
	TDate
	TInt
	TBool
	TEntity
	TList
)

var primitives = map[string]TKind{
	"string": TString, "slug": TSlug, "markdown": TMarkdown,
	"date": TDate, "int": TInt, "bool": TBool,
}

var primitiveNames = []string{"string", "slug", "markdown", "date", "int", "bool"}

type Type struct {
	Kind   TKind
	Entity string
	Elem   *Type
}

func (t *Type) String() string {
	switch t.Kind {
	case TEntity:
		return t.Entity
	case TList:
		return "[" + t.Elem.String() + "]"
	}
	return primitiveNames[t.Kind]
}

func (t *Type) textual() bool   { return t.Kind == TString || t.Kind == TSlug }
func (t *Type) orderable() bool { return t.textual() || t.Kind == TDate || t.Kind == TInt }
func (t *Type) scalar() bool    { return t.Kind <= TBool && t.Kind != TMarkdown }

var targets = []string{"nextjs", "sveltekit"}
var renderers = []string{"markdown", "chips"}
var functions = []string{"excerpt"}

func EntityID(name string) string    { return "entity:" + name }
func PageID(route string) string     { return "page:" + route }
func ComponentID(name string) string { return "component:" + name }

type Program struct {
	File       *File
	Entities   map[string]*Entity
	Components map[string]*Component
	Pages      map[string]*Page
	// Deps maps a node ID to the node IDs whose interfaces it relies on.
	Deps map[string]map[string]bool
}

type checker struct {
	*Program
	d    *Diags
	dir  string
	used map[string]bool
}

// Load runs the whole front end on one source file.
func Load(path, src string) (*Program, *Diags) {
	d := &Diags{}
	toks := lex(src, d)
	p := &parser{toks: toks, d: d}
	f := decode(path, src, p.parseBlocks(), d)
	if d.HasErrors() {
		return nil, d
	}
	prog := Check(f, d)
	if d.HasErrors() {
		return nil, d
	}
	return prog, d
}

func Check(f *File, d *Diags) *Program {
	c := &checker{
		Program: &Program{
			File:       f,
			Entities:   map[string]*Entity{},
			Components: map[string]*Component{},
			Pages:      map[string]*Page{},
			Deps:       map[string]map[string]bool{},
		},
		d:    d,
		dir:  filepath.Dir(f.Path),
		used: map[string]bool{},
	}
	c.collect()
	c.checkApp()
	for _, e := range f.Entities {
		c.checkEntity(e)
	}
	for _, comp := range f.Components {
		c.checkSignature(comp)
	}
	for _, pg := range f.Pages {
		c.checkPage(pg)
	}
	for _, comp := range f.Components {
		c.checkBody(comp)
	}
	for _, comp := range f.Components {
		if !c.used[comp.Name.Name] {
			d.Warnf(comp.Name.Pos, "component %s isn't used by any page", comp.Name.Name)
		}
	}
	return c.Program
}

func (c *checker) dep(from, to string) {
	if from == to {
		return
	}
	if c.Deps[from] == nil {
		c.Deps[from] = map[string]bool{}
	}
	c.Deps[from][to] = true
}

func (c *checker) collect() {
	types := map[string]Pos{}
	claim := func(n Ident) bool {
		if prev, ok := types[n.Name]; ok {
			c.d.Errorf(n.Pos, "'%s' is already declared", n.Name)
			c.d.Hint("first declared on line %d", prev.Line)
			return false
		}
		if _, ok := primitives[n.Name]; ok {
			c.d.Errorf(n.Pos, "'%s' is a built-in type name", n.Name)
			return false
		}
		types[n.Name] = n.Pos
		return true
	}
	for _, e := range c.File.Entities {
		if claim(e.Name) {
			c.Entities[e.Name.Name] = e
			c.Deps[EntityID(e.Name.Name)] = map[string]bool{}
		}
	}
	for _, comp := range c.File.Components {
		if claim(comp.Name) {
			c.Components[comp.Name.Name] = comp
			c.Deps[ComponentID(comp.Name.Name)] = map[string]bool{}
		}
	}
	shapes := map[string]*Page{}
	for _, pg := range c.File.Pages {
		// "/blog/[slug]" and "/blog/[id]" would both match the same URLs.
		var shape []string
		for _, s := range pg.Route.Segs {
			if s.Param != "" {
				shape = append(shape, "[]")
			} else {
				shape = append(shape, s.Lit)
			}
		}
		key := "/" + strings.Join(shape, "/")
		if prev, ok := shapes[key]; ok {
			c.d.Errorf(pg.Route.Pos, "route %q matches the same URLs as %q", pg.Route.Raw, prev.Route.Raw)
			c.d.Hint("declared on line %d", prev.Pos.Line)
			continue
		}
		shapes[key] = pg
		c.Pages[pg.Route.Raw] = pg
		c.Deps[PageID(pg.Route.Raw)] = map[string]bool{}
	}
}

func (c *checker) sortedEntityNames() []string {
	var out []string
	for n := range c.Entities {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func (c *checker) checkApp() {
	a := c.File.App
	if a.Target != nil && !contains(targets, a.Target.Name) {
		c.d.Errorf(a.Target.Pos, "unknown target '%s'", a.Target.Name)
		c.d.hintNames(a.Target.Name, targets, "targets")
	}
}

// resolveType turns a type reference into a Type, reporting unknown names.
func (c *checker) resolveType(tr TypeRef) *Type {
	var t *Type
	if k, ok := primitives[tr.Name]; ok {
		t = &Type{Kind: k}
	} else if _, ok := c.Entities[tr.Name]; ok {
		t = &Type{Kind: TEntity, Entity: tr.Name}
	} else {
		c.d.Errorf(tr.Pos, "unknown type '%s'", tr.Name)
		c.d.hintNames(tr.Name, append(append([]string{}, primitiveNames...), c.sortedEntityNames()...), "types")
		return nil
	}
	if tr.List {
		return &Type{Kind: TList, Elem: t}
	}
	return t
}

// fieldType resolves silently; bad field types were reported by checkEntity.
func (c *checker) fieldType(f *Field) *Type {
	var t *Type
	if k, ok := primitives[f.Type.Name]; ok {
		t = &Type{Kind: k}
	} else if _, ok := c.Entities[f.Type.Name]; ok {
		t = &Type{Kind: TEntity, Entity: f.Type.Name}
	} else {
		return nil
	}
	if f.Type.List {
		return &Type{Kind: TList, Elem: t}
	}
	return t
}

func isUnique(f *Field) bool { return f.Unique || (f.Type.Name == "slug" && !f.Type.List) }

func (c *checker) checkEntity(e *Entity) {
	id := EntityID(e.Name.Name)
	if len(e.Fields) == 0 {
		c.d.Errorf(e.Name.Pos, "entity %s has no fields", e.Name.Name)
	}
	seen := map[string]bool{}
	for _, f := range e.Fields {
		if seen[f.Name.Name] {
			c.d.Errorf(f.Name.Pos, "%s.%s is declared twice", e.Name.Name, f.Name.Name)
		}
		seen[f.Name.Name] = true
		if f.Name.Name == "id" {
			c.d.Errorf(f.Name.Pos, "'id' is reserved; every entity gets one automatically")
		}
		t := c.resolveType(f.Type)
		if t == nil {
			continue
		}
		elem := t
		if t.Kind == TList {
			elem = t.Elem
		}
		if elem.Kind == TEntity {
			c.dep(id, EntityID(elem.Entity))
		}
		if f.Max != nil {
			if t.Kind != TString {
				c.d.Errorf(f.Max.Pos, "(max N) only applies to string fields, and %s.%s is %s", e.Name.Name, f.Name.Name, t)
			} else if f.Max.Value <= 0 {
				c.d.Errorf(f.Max.Pos, "max length must be positive")
			}
		}
		if f.SlugFrom != nil {
			src := e.Field(f.SlugFrom.Name)
			if src == nil {
				c.d.Errorf(f.SlugFrom.Pos, "%s has no field '%s'", e.Name.Name, f.SlugFrom.Name)
				c.d.hintNames(f.SlugFrom.Name, e.FieldNames(), "fields of "+e.Name.Name)
			} else if st := c.fieldType(src); st != nil && st.Kind != TString {
				c.d.Errorf(f.SlugFrom.Pos, "a slug can only be made from a string field, and %s is %s", f.SlugFrom.Name, st)
			}
		}
		if f.Unique && !t.scalar() {
			c.d.Errorf(f.Type.Pos, "%s fields can't be unique", t)
		}
	}
}

func (c *checker) checkSignature(comp *Component) {
	id := ComponentID(comp.Name.Name)
	seen := map[string]bool{}
	for _, p := range comp.Params {
		if seen[p.Name.Name] {
			c.d.Errorf(p.Name.Pos, "parameter '%s' is declared twice", p.Name.Name)
		}
		seen[p.Name.Name] = true
		if t := c.resolveType(p.Type); t != nil {
			if t.Kind == TEntity {
				c.dep(id, EntityID(t.Entity))
			} else if t.Kind == TList && t.Elem.Kind == TEntity {
				c.dep(id, EntityID(t.Elem.Entity))
			}
		}
	}
	if comp.Native && comp.From != nil {
		if _, err := os.Stat(filepath.Join(c.dir, comp.From.Value)); err != nil {
			c.d.Errorf(comp.From.Pos, "native file %q not found", comp.From.Value)
		}
	}
}

func (c *checker) lookupEntity(n Ident) *Entity {
	e := c.Entities[n.Name]
	if e == nil {
		c.d.Errorf(n.Pos, "unknown entity '%s'", n.Name)
		c.d.hintNames(n.Name, c.sortedEntityNames(), "entities")
	}
	return e
}

func (c *checker) lookupField(e *Entity, n Ident) *Field {
	f := e.Field(n.Name)
	if f == nil {
		c.d.Errorf(n.Pos, "%s has no field '%s'", e.Name.Name, n.Name)
		c.d.hintNames(n.Name, e.FieldNames(), "fields of "+e.Name.Name)
	}
	return f
}

func (c *checker) checkPage(pg *Page) {
	id := PageID(pg.Route.Raw)
	q := pg.Data
	if q == nil {
		return
	}
	e := c.lookupEntity(q.Entity)
	if e == nil {
		return
	}
	en := e.Name.Name
	c.dep(id, EntityID(en))
	single := q.By != nil

	if single {
		if f := c.lookupField(e, *q.By); f != nil && !isUnique(f) {
			c.d.Errorf(q.By.Pos, "%s.%s isn't unique, so a lookup by it could match several records", en, f.Name.Name)
			c.d.Hint("mark it `unique` in entity %s, or look up by a unique field", en)
		}
		found := false
		for _, s := range pg.Route.Segs {
			if s.Param == "" {
				continue
			}
			if s.Param == q.By.Name {
				found = true
				continue
			}
			c.d.Errorf(s.Pos, "route param [%s] isn't used by data", s.Param)
			c.d.Hint("data looks up %s by %s, so the param should be [%s]", en, q.By.Name, q.By.Name)
		}
		if !found {
			c.d.Errorf(q.By.Pos, "data looks up %s by %s, but route %q has no [%s] param", en, q.By.Name, pg.Route.Raw, q.By.Name)
		}
	} else {
		for _, s := range pg.Route.Segs {
			if s.Param != "" {
				c.d.Errorf(s.Pos, "route param [%s] isn't used by data", s.Param)
				c.d.Hint("to show one record, write `data: %s by %s`", en, s.Param)
			}
		}
		for _, cond := range q.Where {
			c.checkCond(e, cond)
		}
		if q.Order != nil {
			if f := c.lookupField(e, q.Order.Field); f != nil {
				if t := c.fieldType(f); t != nil && !t.orderable() {
					c.d.Errorf(q.Order.Field.Pos, "can't order by %s.%s: %s values have no order", en, f.Name.Name, t)
				}
			}
		}
		if q.Limit != nil && q.Limit.Value <= 0 {
			c.d.Errorf(q.Limit.Pos, "limit must be positive")
		}
	}

	if l := pg.Layout; l != nil {
		comp := c.Components[l.Component.Name]
		if comp == nil {
			c.d.Errorf(l.Component.Pos, "unknown component '%s'", l.Component.Name)
			var names []string
			for _, cc := range c.File.Components {
				names = append(names, cc.Name.Name)
			}
			c.d.hintNames(l.Component.Name, names, "components")
		} else {
			c.used[comp.Name.Name] = true
			c.dep(id, ComponentID(comp.Name.Name))
			switch {
			case l.List && single:
				c.d.Errorf(l.Pos, "`list of %s` needs a list, but data looks up a single %s", comp.Name.Name, en)
				c.d.Hint("write `layout: %s`", comp.Name.Name)
			case !l.List && !single:
				c.d.Errorf(l.Pos, "data returns a list of %s, so the layout must be a list", en)
				c.d.Hint("write `layout: list of %s`", comp.Name.Name)
			}
			if len(comp.Params) != 1 {
				c.d.Errorf(l.Component.Pos, "%s takes %d parameters; a layout component must take exactly one", comp.Name.Name, len(comp.Params))
			} else if pt := comp.Params[0].Type; pt.List || pt.Name != en {
				c.d.Errorf(l.Component.Pos, "%s expects a %s, but this page gives it a %s", comp.Name.Name, typeRefString(pt), en)
			}
		}
	}

	if single {
		if pg.Empty != nil {
			c.d.Errorf(pg.Empty.Pos, "'empty' only applies to list pages")
		}
		if pg.NotFound == nil {
			c.d.Errorf(pg.Route.Pos, "page %q looks up one %s but doesn't say what happens when it's missing", pg.Route.Raw, en)
			c.d.Hint("add `not_found: 404`")
		} else if pg.NotFound.Value != 404 {
			c.d.Errorf(pg.NotFound.Pos, "not_found only supports 404 for now")
		}
	} else {
		if pg.NotFound != nil {
			c.d.Errorf(pg.NotFound.Pos, "'not_found' only applies to pages that look up one record")
		}
		if pg.Empty == nil {
			c.d.Errorf(pg.Route.Pos, "list page %q doesn't say what to show when there are no %s records", pg.Route.Raw, en)
			c.d.Hint("add `empty: \"No %ss yet.\"`", strings.ToLower(en))
		}
	}
}

func (c *checker) checkCond(e *Entity, cond *Cond) {
	f := c.lookupField(e, cond.Field)
	if f == nil {
		return
	}
	ft := c.fieldType(f)
	if ft == nil {
		return
	}
	if !ft.scalar() {
		c.d.Errorf(cond.Field.Pos, "can't filter on %s.%s: %s values can't be compared", e.Name.Name, f.Name.Name, ft)
		return
	}
	var vt *Type
	switch v := cond.Value.(type) {
	case *IntLit:
		vt = &Type{Kind: TInt}
	case *StrLit:
		vt = &Type{Kind: TString}
	case *PathExpr:
		if len(v.Parts) == 1 {
			switch v.Parts[0].Name {
			case "now":
				vt = &Type{Kind: TDate}
			case "true", "false":
				vt = &Type{Kind: TBool}
			}
		}
	}
	if vt == nil {
		c.d.Errorf(cond.Value.exprPos(), "a condition can only compare against now, true, false, a number or a string")
		return
	}
	if !(ft.Kind == vt.Kind || (ft.textual() && vt.textual())) {
		c.d.Errorf(cond.Value.exprPos(), "can't compare %s.%s (%s) with a %s", e.Name.Name, f.Name.Name, ft, vt)
		return
	}
	if cond.Op != "==" && cond.Op != "!=" && !ft.orderable() {
		c.d.Errorf(cond.OpPos, "%s values can't be compared with %s", ft, cond.Op)
	}
}

// pathRes is what a field path resolves to: its type, and if it ends on an
// entity field, which one.
type pathRes struct {
	t      *Type
	entity string
	field  string
}

func (c *checker) pathType(owner string, scope map[string]*Type, p *PathExpr) *pathRes {
	head := p.Parts[0]
	t := scope[head.Name]
	if t == nil {
		c.d.Errorf(head.Pos, "unknown name '%s'", head.Name)
		var names []string
		for n := range scope {
			names = append(names, n)
		}
		sort.Strings(names)
		c.d.hintNames(head.Name, names, "parameters in scope")
		return nil
	}
	r := &pathRes{t: t}
	for _, part := range p.Parts[1:] {
		if r.t.Kind != TEntity {
			c.d.Errorf(part.Pos, "can't read .%s from a %s", part.Name, r.t)
			if r.t.Kind == TList {
				c.d.Hint("this is a list; show it whole, e.g. `as chips`")
			}
			return nil
		}
		e := c.Entities[r.t.Entity]
		c.dep(owner, EntityID(e.Name.Name))
		f := c.lookupField(e, part)
		if f == nil {
			return nil
		}
		ft := c.fieldType(f)
		if ft == nil {
			return nil
		}
		r = &pathRes{t: ft, entity: e.Name.Name, field: f.Name.Name}
	}
	return r
}

func (c *checker) exprType(owner string, scope map[string]*Type, x Expr) *Type {
	switch x := x.(type) {
	case *IntLit:
		return &Type{Kind: TInt}
	case *StrLit:
		if len(x.Value) > maxCopyLen {
			c.d.Errorf(x.Pos, "string is %d characters; UI copy is capped at %d", len(x.Value), maxCopyLen)
		}
		return &Type{Kind: TString}
	case *PathExpr:
		if r := c.pathType(owner, scope, x); r != nil {
			return r.t
		}
		return nil
	case *CallExpr:
		switch x.Fn.Name {
		case "excerpt":
			if len(x.Args) != 2 {
				c.d.Errorf(x.Fn.Pos, "excerpt takes 2 arguments (text, length), got %d", len(x.Args))
				return nil
			}
			if t := c.exprType(owner, scope, x.Args[0]); t != nil && t.Kind != TMarkdown && t.Kind != TString {
				c.d.Errorf(x.Args[0].exprPos(), "excerpt needs markdown or string text, got %s", t)
			}
			if n, ok := x.Args[1].(*IntLit); !ok || n.Value <= 0 {
				c.d.Errorf(x.Args[1].exprPos(), "excerpt length must be a positive number")
			}
			return &Type{Kind: TString}
		default:
			c.d.Errorf(x.Fn.Pos, "unknown function '%s'", x.Fn.Name)
			c.d.hintNames(x.Fn.Name, functions, "functions")
			return nil
		}
	}
	return nil
}

// displayFields are the fields a chip could show for an entity.
func displayFields(e *Entity) []string {
	var out []string
	for _, f := range e.Fields {
		if f.Type.Name == "string" && !f.Type.List {
			out = append(out, f.Name.Name)
		}
	}
	return out
}

func (c *checker) checkBody(comp *Component) {
	if comp.Native {
		return
	}
	id := ComponentID(comp.Name.Name)
	scope := map[string]*Type{}
	for _, p := range comp.Params {
		if t := c.fieldType(&Field{Type: p.Type}); t != nil {
			scope[p.Name.Name] = t
		}
	}
	for _, s := range comp.Shows {
		t := c.exprType(id, scope, s.Expr)
		if t == nil {
			continue
		}
		c.checkRender(id, s, t)
	}
	if comp.LinksTo != nil {
		c.checkLink(id, scope, comp.LinksTo)
	}
}

func (c *checker) checkRender(owner string, s *Show, t *Type) {
	pos := s.Expr.exprPos()
	if s.As == nil {
		switch t.Kind {
		case TMarkdown:
			c.d.Errorf(pos, "markdown can't be shown as plain text")
			c.d.Hint("write `%s as markdown`, or `excerpt(%s, 160)`", exprString(s.Expr), exprString(s.Expr))
		case TEntity:
			c.d.Errorf(pos, "say which field of %s to show", t.Entity)
			if fs := c.Entities[t.Entity].FieldNames(); len(fs) > 0 {
				c.d.Hint("e.g. %s.%s", exprString(s.Expr), fs[0])
			}
		case TList:
			c.d.Errorf(pos, "say how to show this list")
			c.d.Hint("e.g. `%s as chips`", exprString(s.Expr))
		}
		return
	}
	switch s.As.Name {
	case "markdown":
		if t.Kind != TMarkdown {
			c.d.Errorf(s.As.Pos, "`as markdown` needs a markdown value, got %s", t)
		}
	case "chips":
		if t.Kind != TList {
			c.d.Errorf(s.As.Pos, "`as chips` needs a list, got %s", t)
			return
		}
		switch el := t.Elem; {
		case el.Kind == TEntity:
			e := c.Entities[el.Entity]
			c.dep(owner, EntityID(e.Name.Name))
			if df := displayFields(e); len(df) != 1 {
				c.d.Errorf(s.As.Pos, "can't tell which field of %s a chip should show", e.Name.Name)
				if len(df) == 0 {
					c.d.Hint("%s has no string field", e.Name.Name)
				} else {
					c.d.Hint("%s has several string fields: %s", e.Name.Name, strings.Join(df, ", "))
				}
			}
		case !el.textual():
			c.d.Errorf(s.As.Pos, "`as chips` needs a list of text or records, got %s", t)
		}
	default:
		c.d.Errorf(s.As.Pos, "unknown renderer '%s'", s.As.Name)
		c.d.hintNames(s.As.Name, renderers, "renderers")
	}
}

func (c *checker) checkLink(owner string, scope map[string]*Type, tpl *Template) {
	res := make([]*pathRes, len(tpl.Segs))
	for i, s := range tpl.Segs {
		if s.Path == nil {
			continue
		}
		r := c.pathType(owner, scope, s.Path)
		if r == nil {
			return
		}
		if !r.t.scalar() {
			c.d.Errorf(s.Pos, "a %s can't go in a URL", r.t)
			return
		}
		res[i] = r
	}

	var target *Page
	for _, pg := range c.File.Pages {
		if len(pg.Route.Segs) != len(tpl.Segs) {
			continue
		}
		match := true
		for i, s := range pg.Route.Segs {
			if (s.Param != "") != (tpl.Segs[i].Path != nil) || s.Lit != tpl.Segs[i].Lit {
				match = false
				break
			}
		}
		if match {
			target = pg
			break
		}
	}
	if target == nil {
		c.d.Errorf(tpl.Pos, "no page matches the link %q", tpl.Raw)
		var routes []string
		for _, pg := range c.File.Pages {
			routes = append(routes, pg.Route.Raw)
		}
		c.d.Hint("pages: %s", strings.Join(routes, ", "))
		return
	}
	c.dep(owner, PageID(target.Route.Raw))
	q := target.Data
	if q == nil || q.By == nil {
		return
	}
	for i, s := range tpl.Segs {
		if s.Path == nil {
			continue
		}
		r := res[i]
		if r.entity == q.Entity.Name && r.field == q.By.Name {
			continue
		}
		got := exprString(s.Path)
		if r.entity != "" {
			got = r.entity + "." + r.field
		}
		c.d.Errorf(s.Pos, "page %q looks up %s by %s, but this link passes %s", target.Route.Raw, q.Entity.Name, q.By.Name, got)
		for name, t := range scope {
			if t.Kind == TEntity && t.Entity == q.Entity.Name {
				c.d.Hint("use [%s.%s]", name, q.By.Name)
				break
			}
		}
	}
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
