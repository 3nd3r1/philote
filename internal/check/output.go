package check

import (
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

// output is one value an endpoint returns by name, and where it was written.
type output struct {
	value *ir.Value
	field *syntax.Field
}

// outputs resolves `output: { ... }`, whose values may be optional records.
func (c *checker) outputs(e *syntax.Endpoint) []output {
	var out []output
	seen := map[string]bool{}
	for _, f := range e.Outputs {
		if seen[f.Name.Name] {
			c.d.Errorf(f.Name.Pos, "'%s' is listed twice", f.Name.Name)
			continue
		}
		seen[f.Name.Name] = true
		if v := c.value(f); v != nil {
			out = append(out, output{v, f})
		}
	}
	return out
}

// named checks that each value the endpoint returns by name is an input or
// a step's result, of the same type, so the handler only passes it on.
func (c *checker) named(ep *ir.Endpoint, outputs []output, scope map[string]*ir.Value) {
	if scope == nil {
		return
	}
	for _, o := range outputs {
		v := scope[o.value.Name]
		switch {
		case v == nil:
			c.d.Errorf(o.field.Name.Pos, "no input or step gives '%s'", o.value.Name)
			c.d.Hint("available here: %s", strings.Join(sortedKeys(scope), ", "))
		case !sameType(v, o.value):
			c.d.Errorf(o.field.Type.Pos, "'%s' is a %s here, not a %s", o.value.Name, typeName(v), typeName(o.value))
		default:
			ep.Outputs = append(ep.Outputs, o.value)
		}
	}
}
