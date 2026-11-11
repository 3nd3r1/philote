package check

import (
	"slices"
	"sort"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

var partKinds = map[string]ir.PartKind{"backend": ir.Backend, "frontend": ir.Frontend}

func (c *checker) parts(app *syntax.App) {
	seen := map[string]bool{}
	for _, it := range app.Items {
		p, ok := it.(*syntax.Part)
		if !ok {
			c.d.Errorf(it.Start(), "only backend and frontend parts go inside app")
			continue
		}
		if seen[p.Name.Name] {
			c.d.Errorf(p.Name.Pos, "there is already a part named %s", p.Name.Name)
			continue
		}
		seen[p.Name.Name] = true
		if part := c.part(p); part != nil {
			c.app.Parts = append(c.app.Parts, part)
		}
	}
	c.resolveFrontends()
}

func (c *checker) targetNames() []string {
	var out []string
	for n := range c.targets {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func (c *checker) part(p *syntax.Part) *ir.Part {
	kind := partKinds[p.Kind.Name]
	tk, ok := c.targets[p.Target.Name]
	switch {
	case !ok:
		c.d.Errorf(p.Target.Pos, "unknown target '%s'", p.Target.Name)
		c.d.HintNames(p.Target.Name, c.targetNames(), "available targets")
		return nil
	case tk != kind:
		c.d.Errorf(p.Target.Pos, "%s can't build a %s", p.Target.Name, p.Kind.Name)
		return nil
	}
	part := &ir.Part{Name: p.Name.Name, Kind: kind, Target: p.Target.Name}
	if kind == ir.Frontend {
		c.frontend(part, p)
		return part
	}
	var endpoints []*syntax.Endpoint
	for _, it := range p.Body {
		switch x := it.(type) {
		case *syntax.Resource:
			c.resource(part, x)
		case *syntax.Step:
			c.step(part, x)
		case *syntax.Endpoint:
			endpoints = append(endpoints, x)
		default:
			c.d.Errorf(it.Start(), "this can't go inside a %s", p.Kind.Name)
		}
	}
	// Endpoints refer to steps declared anywhere in the part.
	for _, e := range endpoints {
		c.endpoint(part, e)
	}
	c.unusedSteps(part)
	return part
}

func (c *checker) unusedSteps(part *ir.Part) {
	for _, s := range part.Steps {
		used := slices.ContainsFunc(part.Custom, func(e *ir.Endpoint) bool { return slices.Contains(e.Flow.Steps, s) })
		if !used {
			c.d.Warnf(s.Pos, "step %s isn't used by any endpoint", s.Name)
		}
	}
}

// isURLName accepts lowercase names like order_lines: a letter, then letters, digits or '_'.
func isURLName(s string) bool {
	for i := 0; i < len(s); i++ {
		ch := s[i]
		letter := ch >= 'a' && ch <= 'z'
		later := i > 0 && (ch == '_' || (ch >= '0' && ch <= '9'))
		if !letter && !later {
			return false
		}
	}
	return s != ""
}

func (c *checker) resource(part *ir.Part, r *syntax.Resource) {
	if !isURLName(r.Name.Name) {
		c.d.Errorf(r.Name.Pos, "resource names are used in URLs, so they must be lowercase, like `orders`")
		return
	}
	if slices.ContainsFunc(part.Resources, func(x *ir.Resource) bool { return x.Name == r.Name.Name }) {
		c.d.Errorf(r.Name.Pos, "%s already has a resource named %s", part.Name, r.Name.Name)
		return
	}
	e := c.lookupEntity(r.Entity)
	if e == nil {
		return
	}
	res := &ir.Resource{Name: r.Name.Name, Entity: e.Name}
	for _, op := range r.Only {
		i := slices.Index(ir.OpNames[:], op.Name)
		switch {
		case i < 0:
			c.d.Errorf(op.Pos, "unknown operation '%s'", op.Name)
			c.d.HintNames(op.Name, ir.OpNames[:], "operations")
		case slices.Contains(res.Ops, ir.Op(i)):
			c.d.Errorf(op.Pos, "'%s' is listed twice", op.Name)
		default:
			res.Ops = append(res.Ops, ir.Op(i))
		}
	}
	if len(r.Only) == 0 {
		res.Ops = []ir.Op{ir.List, ir.Get, ir.Create, ir.Update, ir.Delete}
	}
	slices.Sort(res.Ops)
	part.Resources = append(part.Resources, res)
}
