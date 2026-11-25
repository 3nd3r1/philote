package check

import (
	"strings"
	"testing"

	"github.com/3nd3r1/philote/internal/diag"
)

// namedApp is callApp where rate returns what it did by name: the review,
// and the product it may have featured.
var namedApp = strings.NewReplacer(
	"      output: Review\n      steps: rate_product\n", "      output: { review: Review, featured: Product optional }\n      steps: rate_product, feature_if_loved\n",
	"      intent: \"store a review of other by product\"\n    }\n", `      intent: "store a review of other by product"
    }

    step feature_if_loved(review: Review) -> featured: Product optional {
      writes: Product
      intent: "feature the product if the review has five stars, and return it; else null"
    }
`,
).Replace(callApp)

func TestNamedOutputs(t *testing.T) {
	app, d := load(map[string]string{"app.philote": namedApp})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": namedApp}))
	}
	rate := app.Parts[0].Custom[1]
	if rate.Output != "" || len(rate.Outputs) != 2 {
		t.Fatalf("rate should return two named values, got output %q and %v", rate.Output, rate.Outputs)
	}
	review, featured := rate.Outputs[0], rate.Outputs[1]
	if review.Name != "review" || review.Entity != "Review" || review.Optional ||
		featured.Name != "featured" || featured.Entity != "Product" || !featured.Optional {
		t.Errorf("want review: Review and featured: Product optional, got %+v and %+v", review, featured)
	}
	if st := app.Parts[0].Steps[2]; st.Result == nil || !st.Result.Optional {
		t.Errorf("feature_if_loved should return an optional Product, got %+v", st.Result)
	}
}

func TestNamedOutputErrors(t *testing.T) {
	runErrorCases(t, namedApp, []errorCase{
		{"names nothing", "{ review: Review,", "{ verdict: Review,", "{ »verdict", "no input or step gives 'verdict'"},
		{"wrong type", "featured: Product optional }", "featured: Product }", "featured: »Product }", "'featured' is a Product optional here, not a Product"},
		{"listed twice", "featured: Product optional }", "review: Review }", "Review, »review: Review }", "'review' is listed twice"},
		{"optional scalar", "liked: bool }", "liked: bool optional }", "bool »optional }", "only one record can be optional"},
		{"optional list", "-> featured: Product optional", "-> featured: list of Product optional", "Product »optional {", "only one record can be optional"},
		{"optional input", "input: { product: Product,", "input: { product: Product optional,", "Product »optional,", "only a step's result or an endpoint's output can be optional"},
		{"optional into a step", "step feature_if_loved(review: Review)", "step feature_if_loved(review: Review optional)", "Review »optional)", "only a step's result or an endpoint's output can be optional"},
		{"optional result into a step", "-> review: Review {", "-> review: Review optional {", "rate_product, »feature_if_loved", "'review' is a Review optional here, but step feature_if_loved takes a Review"},
	})
	// A plain output can't come from an optional result: the client would
	// get null where it expects a record.
	runErrorCases(t, callApp, []errorCase{
		{"optional for one output", "-> review: Review {", "-> review: Review optional {", "output: »Review", "no step returns a Review"},
	})
}
