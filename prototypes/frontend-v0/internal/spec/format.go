package spec

import (
	"fmt"
	"strings"
)

// Canonical printing. The build graph hashes these strings, so formatting,
// comments and whitespace in the source never change a cache key.

func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

func typeRefString(t TypeRef) string {
	if t.List {
		return "[" + t.Name + "]"
	}
	return t.Name
}

func exprString(x Expr) string {
	switch x := x.(type) {
	case *IntLit:
		return fmt.Sprint(x.Value)
	case *StrLit:
		return quote(x.Value)
	case *PathExpr:
		var parts []string
		for _, p := range x.Parts {
			parts = append(parts, p.Name)
		}
		return strings.Join(parts, ".")
	case *CallExpr:
		var args []string
		for _, a := range x.Args {
			args = append(args, exprString(a))
		}
		return x.Fn.Name + "(" + strings.Join(args, ", ") + ")"
	}
	return "?"
}

func paramsString(ps []*Param) string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Name.Name+": "+typeRefString(p.Type))
	}
	return "(" + strings.Join(out, ", ") + ")"
}

func FormatApp(a *App) string {
	var b strings.Builder
	fmt.Fprintf(&b, "app %s {\n", a.Name.Name)
	if a.Target != nil {
		fmt.Fprintf(&b, "  target: %s\n", a.Target.Name)
	}
	if a.Style != nil {
		fmt.Fprintf(&b, "  style: %s\n", a.Style.Value)
	}
	b.WriteString("}\n")
	return b.String()
}

func formatField(f *Field) string {
	s := f.Name.Name + ": " + typeRefString(f.Type)
	if f.SlugFrom != nil {
		s += " from " + f.SlugFrom.Name
	}
	if f.Max != nil {
		s += fmt.Sprintf("(max %d)", f.Max.Value)
	}
	if f.Unique {
		s += " unique"
	}
	if f.Optional {
		s += " optional"
	}
	return s
}

func FormatEntity(e *Entity) string {
	var b strings.Builder
	fmt.Fprintf(&b, "entity %s {\n", e.Name.Name)
	for _, f := range e.Fields {
		fmt.Fprintf(&b, "  %s\n", formatField(f))
	}
	b.WriteString("}\n")
	return b.String()
}

func formatQuery(q *Query) string {
	s := q.Entity.Name
	if q.By != nil {
		return s + " by " + q.By.Name
	}
	for i, c := range q.Where {
		kw := " and "
		if i == 0 {
			kw = " where "
		}
		s += kw + c.Field.Name + " " + c.Op + " " + exprString(c.Value)
	}
	if q.Order != nil {
		s += " order by " + q.Order.Field.Name
		if q.Order.Desc {
			s += " desc"
		}
	}
	if q.Limit != nil {
		s += fmt.Sprintf(" limit %d", q.Limit.Value)
	}
	return s
}

func FormatPage(p *Page) string {
	var b strings.Builder
	fmt.Fprintf(&b, "page %s {\n", quote(p.Route.Raw))
	if p.Data != nil {
		fmt.Fprintf(&b, "  data: %s\n", formatQuery(p.Data))
	}
	if p.Layout != nil {
		l := p.Layout.Component.Name
		if p.Layout.List {
			l = "list of " + l
		}
		fmt.Fprintf(&b, "  layout: %s\n", l)
	}
	if p.Empty != nil {
		fmt.Fprintf(&b, "  empty: %s\n", quote(p.Empty.Value))
	}
	if p.NotFound != nil {
		fmt.Fprintf(&b, "  not_found: %d\n", p.NotFound.Value)
	}
	b.WriteString("}\n")
	return b.String()
}

func FormatComponent(c *Component) string {
	if c.Native {
		return fmt.Sprintf("native component %s%s from %s\n", c.Name.Name, paramsString(c.Params), quote(c.From.Value))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "component %s%s {\n", c.Name.Name, paramsString(c.Params))
	if len(c.Shows) > 0 {
		var items []string
		for _, s := range c.Shows {
			it := exprString(s.Expr)
			if s.As != nil {
				it += " as " + s.As.Name
			}
			items = append(items, it)
		}
		fmt.Fprintf(&b, "  shows: %s\n", strings.Join(items, ", "))
	}
	if c.LinksTo != nil {
		fmt.Fprintf(&b, "  links_to: %s\n", quote(c.LinksTo.Raw))
	}
	b.WriteString("}\n")
	return b.String()
}

func FormatFile(f *File) string {
	var parts []string
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *App:
			parts = append(parts, FormatApp(d))
		case *Entity:
			parts = append(parts, FormatEntity(d))
		case *Page:
			parts = append(parts, FormatPage(d))
		case *Component:
			parts = append(parts, FormatComponent(d))
		}
	}
	return strings.Join(parts, "\n")
}

// Interfaces: the part of a node other nodes can see. Changing a node's body
// without changing its interface must not invalidate its dependents.

func IfaceEntity(e *Entity) string { return FormatEntity(e) }

func IfaceComponent(c *Component) string {
	return "component " + c.Name.Name + paramsString(c.Params)
}

func IfacePage(p *Page) string {
	s := "page " + quote(p.Route.Raw)
	if p.Data != nil {
		if p.Data.By != nil {
			s += " -> " + p.Data.Entity.Name + " by " + p.Data.By.Name
		} else {
			s += " -> [" + p.Data.Entity.Name + "]"
		}
	}
	return s
}
