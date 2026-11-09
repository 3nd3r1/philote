package check

import (
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// relation adds both sides of a relation. The many side gets a list, the
// one side gets a reference and a foreign key named <name>_id.
func (c *checker) relation(r *syntax.Relation) {
	from, to := c.lookupEntity(r.From), c.lookupEntity(r.To)
	if from == nil || to == nil {
		return
	}
	backName, backAt, implicit := snake(r.From.Name), r.To, true
	if r.Back != nil {
		backName, backAt, implicit = r.Back.Name, *r.Back, false
	}
	desc := "the relation at " + r.From.Pos.String()

	if r.Verb.Name == "has many" {
		c.addSide(from, &ir.Relation{Name: r.Name.Name, Many: true, Target: to.Name}, r.Name, desc, false)
		c.addSide(to, &ir.Relation{Name: backName, Target: from.Name, FK: backName + "_id"}, backAt, desc, implicit)
		return
	}
	c.addSide(from, &ir.Relation{Name: r.Name.Name, Target: to.Name, FK: r.Name.Name + "_id"}, r.Name, desc, false)
	c.addSide(to, &ir.Relation{Name: backName, Many: true, Target: from.Name}, backAt, desc, implicit)
}

func (c *checker) addSide(e *ir.Entity, rel *ir.Relation, at syntax.Ident, desc string, implicit bool) {
	ok := c.claimAt(e.Name, rel.Name, at, desc)
	if !ok && implicit {
		c.d.Hint("the reverse side is named %s.%s by default; name it with `(back <name>)`, or remove the duplicate relation", e.Name, rel.Name)
	}
	if ok && rel.FK != "" {
		ok = c.claimAt(e.Name, rel.FK, at, desc)
	}
	if ok {
		e.Relations = append(e.Relations, rel)
	}
}
