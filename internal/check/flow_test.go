package check

import (
	"slices"
	"strings"
	"testing"

	"github.com/3nd3r1/philote/internal/diag"
)

const stepIntent = `      intent: "mark the product featured at that rank"`

func TestStepTablesResolve(t *testing.T) {
	src := strings.Replace(forms, stepIntent, "      reads: Review\n      writes: Product\n"+stepIntent, 1)
	app, d := load(map[string]string{"app.philote": src})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": src}))
	}
	st := app.Parts[0].Steps[0]
	if !slices.Equal(st.Reads, []string{"Review"}) || !slices.Equal(st.Writes, []string{"Product"}) {
		t.Errorf("reads %v, writes %v; want [Review] and [Product]", st.Reads, st.Writes)
	}
}

func TestStepTableErrors(t *testing.T) {
	runErrorCases(t, forms, []errorCase{
		{"unknown table", stepIntent, "      writes: Prodct\n" + stepIntent, "writes: »Prodct", "did you mean 'Product'"},
		{"table twice", stepIntent, "      reads: Review, Review\n" + stepIntent, "Review, »Review", "listed twice"},
		{"read and written", stepIntent, "      reads: Product\n      writes: Product\n" + stepIntent, "writes: »Product", "under both reads and writes"},
		{"unknown key", stepIntent, "      changes: Product\n" + stepIntent, "»changes: Product", "step keys: fails, reads, writes, intent"},
	})
}
