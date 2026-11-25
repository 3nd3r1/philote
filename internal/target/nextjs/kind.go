package nextjs

import (
	"fmt"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
)

// fieldView describes one shown field or relation in a built-in component
// template. Templates only print these expressions; all formatting choices
// live here.
type showView struct {
	Label     string
	RawExpr   string
	Expr      string
	Kind      string // text, long, slug, email, int, bool, date, datetime, enum, relation, chips
	Optional  bool
	HrefExpr  string // for linked belongs-to relations
	ChipsExpr string // for has-many chips
	ChipLabel string // for has-many chips
	ChipHref  bool
}

type formatNeeds struct {
	date, dateTime, boolFmt bool
}

func (f *formatNeeds) list() []string {
	var out []string
	if f.boolFmt {
		out = append(out, "formatBool")
	}
	if f.date {
		out = append(out, "formatDate")
	}
	if f.dateTime {
		out = append(out, "formatDateTime")
	}
	return out
}

// buildComponentFields splits a component's shows into title, avatar, prose,
// metadata, and chips so each kind template renders without logic.
func buildComponentFields(app *ir.App, c *ir.Component, v *componentView) {
	if c.Param == "" {
		return
	}
	varName := c.Param
	if c.List {
		varName = "item"
	}
	ent := app.Entity(c.Entity)
	var needs formatNeeds
	all := make([]showView, 0, len(c.Shows))
	for _, s := range c.Shows {
		fv := makeShowView(app, varName, s, &needs)
		// A picker tile is already a button, so relations inside it render as
		// plain text rather than nested links.
		if c.Kind == "picker" {
			fv.HrefExpr = ""
		}
		all = append(all, fv)
	}
	v.AllFields = all
	v.FormatImports = strings.Join(needs.list(), ", ")
	if c.Kind == "table" {
		v.Fields = all
		return
	}
	titleIdx := pickTitleIndex(ent, c.Shows)
	title := all[titleIdx]
	v.Title = &title
	if title.Kind == "text" && !title.Optional && (c.Kind == "card" || c.Kind == "detail" || c.Kind == "picker" || c.Kind == "deck") {
		v.AvatarExpr = title.RawExpr
	}
	for i, fv := range all {
		if i == titleIdx {
			continue
		}
		v.Fields = append(v.Fields, fv)
		switch fv.Kind {
		case "long":
			v.LongFields = append(v.LongFields, fv)
		case "chips":
			v.ChipFields = append(v.ChipFields, fv)
		default:
			v.MetaFields = append(v.MetaFields, fv)
		}
	}
}

func pickTitleIndex(ent *ir.Entity, shows []*ir.Show) int {
	if ent.Display != "" {
		for i, s := range shows {
			if s.Field != nil && s.Name == ent.Display {
				return i
			}
		}
	}
	return 0
}

func makeShowView(app *ir.App, varName string, s *ir.Show, needs *formatNeeds) showView {
	lbl := humanLabel(s.Name)
	if s.Relation != nil {
		return makeRelationShowView(app, varName, lbl, s, needs)
	}
	raw := varName + "." + s.Name
	fv := showView{Label: lbl, RawExpr: raw, Optional: s.Field.Optional}
	fv.Kind, fv.Expr = formatScalar(raw, s.Field, needs)
	return fv
}

func makeRelationShowView(app *ir.App, varName, lbl string, s *ir.Show, needs *formatNeeds) showView {
	targetEnt := app.Entity(s.Relation.Target)
	dispField := findField(targetEnt, s.Display)
	if s.Relation.Many {
		_, chipExpr := formatScalar("chip."+s.Display, dispField, needs)
		return showView{
			Label:     lbl,
			Kind:      "chips",
			ChipsExpr: varName + "." + s.Name,
			ChipLabel: chipExpr,
			ChipHref:  s.Href != "",
		}
	}
	raw := varName + "." + s.Name + "." + s.Display
	_, expr := formatScalar(raw, dispField, needs)
	fv := showView{Label: lbl, RawExpr: raw, Expr: expr, Kind: "relation"}
	if s.Href != "" {
		fv.HrefExpr = varName + "." + s.Name + ".href"
	}
	return fv
}

func formatScalar(raw string, f *ir.Field, needs *formatNeeds) (string, string) {
	opt := f.Optional
	switch f.Kind {
	case ir.Date:
		needs.date = true
		return "date", optExpr(opt, raw, fmt.Sprintf("formatDate(%s)", raw))
	case ir.DateTime:
		needs.dateTime = true
		return "datetime", optExpr(opt, raw, fmt.Sprintf("formatDateTime(%s)", raw))
	case ir.Bool:
		needs.boolFmt = true
		if opt {
			return "bool", fmt.Sprintf(`%s != null ? formatBool(%s) : "—"`, raw, raw)
		}
		return "bool", fmt.Sprintf("formatBool(%s)", raw)
	case ir.Int:
		if opt {
			return "int", fmt.Sprintf(`%s != null ? %s.toLocaleString("en") : "—"`, raw, raw)
		}
		return "int", fmt.Sprintf(`%s.toLocaleString("en")`, raw)
	case ir.Slug:
		return "slug", optExpr(opt, raw, "`@${"+raw+"}`")
	case ir.Email:
		return "email", optExpr(opt, raw, raw)
	case ir.EnumValue:
		return "enum", optExpr(opt, raw, raw)
	default:
		kind := "text"
		if f.Max != nil && *f.Max > 140 {
			kind = "long"
		}
		return kind, optExpr(opt, raw, raw)
	}
}

func optExpr(optional bool, raw, formatted string) string {
	if !optional {
		return formatted
	}
	return fmt.Sprintf(`%s ? %s : "—"`, raw, formatted)
}

func findField(ent *ir.Entity, name string) *ir.Field {
	for _, f := range ent.Fields {
		if f.Name == name {
			return f
		}
	}
	panic("nextjs: entity " + ent.Name + " has no field " + name)
}

func humanLabel(name string) string {
	s := strings.ReplaceAll(name, "_", " ")
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func setComponentCalls(backend *ir.Part, c *ir.Component, v *componentView) {
	if c.Calls == nil {
		return
	}
	v.Client = true
	v.Call = camel(c.Calls.Endpoint)
	v.CallFile = kebab(c.Name) + "-call"
	for _, in := range c.Calls.Passed {
		switch {
		case in.Entity == c.Entity && !in.List:
			v.CallRecordArg = in.Name
		case in.Field != nil && in.Field.Kind == ir.Bool:
			v.CallBoolArg = in.Name
		}
	}
	ep := endpoint(backend, c.Calls.Endpoint)
	for _, out := range ep.Outputs {
		if out.Optional {
			v.MatchField = out.Name
			break
		}
	}
}

func setComponentImports(c *ir.Component, v *componentView) {
	hasFieldLink := slices.ContainsFunc(v.AllFields, func(f showView) bool { return f.HrefExpr != "" || f.ChipHref })
	v.ImportLink = c.LinksTo != "" || hasFieldLink
	v.ImportMark = v.AvatarExpr != ""
	v.ImportButton = c.Kind == "deck"
	v.ImportButtonVariants = c.LinkLabel != ""
	v.ImportCard = c.Kind == "card" || c.Kind == "detail" || c.Kind == "deck"
	v.ImportCardContent = (c.Kind == "card" && len(v.Fields) > 0) ||
		(c.Kind == "detail" && (len(v.LongFields) > 0 || len(v.MetaFields) > 0 || len(v.ChipFields) > 0)) ||
		c.Kind == "deck"
	v.ImportCardFooter = (c.Kind == "card" || c.Kind == "detail") && (v.Actions != "" || v.LinkLabel != "")
	v.ImportBadge = slices.ContainsFunc(v.AllFields, func(f showView) bool {
		return f.Kind == "enum" || f.Kind == "bool" || f.Kind == "slug" || f.Kind == "chips"
	})
}
