package check

import (
	"slices"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

var componentKinds = []string{"card", "detail", "table", "list", "hero", "callout", "picker", "deck"}

// componentKindHeader checks the kind name and parameter shape before the
// component's body is resolved, so a wrong parameter shape gets a direct error.
func (c *checker) componentKindHeader(comp *ir.Component, s *syntax.Component) bool {
	if s.Kind == nil {
		return true
	}
	k := s.Kind.Name
	if !slices.Contains(componentKinds, k) {
		c.d.Errorf(s.Kind.Pos, "unknown component kind '%s'", k)
		c.d.HintNames(k, componentKinds, "component kinds")
		return false
	}
	comp.Kind = k
	if s.Look != nil {
		c.d.Errorf(s.Look.Pos, "a %s component is built from a template and doesn't take look", k)
		return false
	}
	return c.kindParam(comp, s)
}

func (c *checker) kindParam(comp *ir.Component, s *syntax.Component) bool {
	hasParam := len(s.Params) == 1
	isList := hasParam && s.Params[0].List
	switch comp.Kind {
	case "card", "detail":
		if !hasParam || isList {
			c.d.Errorf(s.Kind.Pos, "a %s component takes one record, like %s(item: Item)", comp.Kind, comp.Name)
			return false
		}
	case "table", "list", "picker", "deck":
		if !isList {
			c.d.Errorf(s.Kind.Pos, "a %s component takes a list, like %s(items: list of Item)", comp.Kind, comp.Name)
			return false
		}
	case "hero", "callout":
		if hasParam {
			c.d.Errorf(s.Kind.Pos, "a %s component takes no record", comp.Kind)
			return false
		}
	}
	return true
}

// componentKindClauses checks kind-specific rules once shows and calls are
// resolved.
func (c *checker) componentKindClauses(comp *ir.Component, s *syntax.Component) {
	switch comp.Kind {
	case "":
	case "card", "detail":
		c.kindRecord(comp, s)
	case "table", "list":
		c.kindList(comp, s)
	case "hero", "callout":
		c.kindCopy(comp, s)
	case "picker":
		c.kindPicker(comp, s)
	case "deck":
		c.kindDeck(comp, s)
	}
}

func (c *checker) kindRecord(comp *ir.Component, s *syntax.Component) {
	if len(s.Says) > 0 {
		c.d.Errorf(s.Says[0].Pos, "a %s component doesn't take says", comp.Kind)
	}
	if s.Picks != nil {
		c.d.Errorf(s.Picks.Pos, "a %s component doesn't take picks", comp.Kind)
	}
	if s.Calls != nil {
		c.d.Errorf(s.Calls.Endpoint.Pos, "a %s component doesn't take calls", comp.Kind)
	}
	if comp.Kind == "detail" && s.LinksTo != nil && s.LinkLabel == nil {
		c.d.Errorf(s.LinksTo.Pos, "a link on a detail component needs a button text")
		c.d.Hint(`add it after the route: links_to: "%s" as "..."`, s.LinksTo.Value)
	}
}

func (c *checker) kindList(comp *ir.Component, s *syntax.Component) {
	if len(s.Says) > 0 {
		c.d.Errorf(s.Says[0].Pos, "a %s component doesn't take says", comp.Kind)
	}
	if s.Picks != nil {
		c.d.Errorf(s.Picks.Pos, "a %s component doesn't take picks", comp.Kind)
	}
	if s.Calls != nil {
		c.d.Errorf(s.Calls.Endpoint.Pos, "a %s component doesn't take calls", comp.Kind)
	}
}

func (c *checker) kindCopy(comp *ir.Component, s *syntax.Component) {
	maxSays := 2
	if comp.Kind == "callout" {
		maxSays = 1
	}
	if len(s.Says) > maxSays {
		c.d.Errorf(s.Says[maxSays].Pos, "a %s component takes at most %d says", comp.Kind, maxSays)
	}
	if s.Picks != nil {
		c.d.Errorf(s.Picks.Pos, "a %s component doesn't take picks", comp.Kind)
	}
	if s.Calls != nil {
		c.d.Errorf(s.Calls.Endpoint.Pos, "a %s component doesn't take calls", comp.Kind)
	}
}

func (c *checker) kindPicker(comp *ir.Component, s *syntax.Component) {
	if len(s.Says) > 1 {
		c.d.Errorf(s.Says[1].Pos, "a picker component takes at most 1 says")
	}
	switch {
	case s.Picks != nil && s.Calls != nil:
		c.d.Errorf(s.Calls.Endpoint.Pos, "a picker component takes picks or calls, not both")
	case s.Picks == nil && s.Calls == nil:
		c.d.Errorf(s.Kind.Pos, "a picker component needs picks or calls")
		c.d.Hint(`add picks: viewer then "/..." or calls: <endpoint>(...)`)
	case comp.Calls != nil:
		if len(comp.Calls.Passed) != 1 || comp.Calls.Passed[0].Entity != comp.Entity || comp.Calls.Passed[0].List {
			c.d.Errorf(s.Calls.Endpoint.Pos, "a picker's calls passes only the tapped %s record (plus any viewer inputs)", comp.Entity)
		}
	}
}

func (c *checker) kindDeck(comp *ir.Component, s *syntax.Component) {
	switch {
	case len(s.Says) == 0:
		c.d.Errorf(s.Kind.Pos, "a deck component needs says with at least an empty-deck message")
	case len(s.Says) > 3:
		c.d.Errorf(s.Says[3].Pos, "a deck component takes at most 3 says")
	}
	if s.Picks != nil {
		c.d.Errorf(s.Picks.Pos, "a deck component doesn't take picks")
	}
	if s.Calls == nil {
		c.d.Errorf(s.Kind.Pos, "a deck component needs calls for its swipe endpoint")
		return
	}
	if comp.Calls == nil {
		return
	}
	var records, bools int
	for _, v := range comp.Calls.Passed {
		switch {
		case v.Entity == comp.Entity && !v.List:
			records++
		case v.Field != nil && v.Field.Kind == ir.Bool:
			bools++
		}
	}
	if len(comp.Calls.Passed) != 2 || records != 1 || bools != 1 {
		c.d.Errorf(s.Calls.Endpoint.Pos, "a deck's calls passes the current %s record and one bool (plus any viewer inputs)", comp.Entity)
	}
}
