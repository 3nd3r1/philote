package check

import (
	"slices"
	"strings"
	"testing"

	"github.com/3nd3r1/philote/internal/diag"
)

// callApp is the viewer app where /for-me loads the products related to the
// viewer from a GET endpoint, and RateDeck rates one of them as the viewer.
var callApp = strings.NewReplacer(
	"    resource reviews: Review { only: list }\n", `    resource reviews: Review { only: list }

    endpoint related {
      route: GET "/products/[product]/related"
      input: { product: Product }
      output: list of Product
      steps: find_related
    }

    step find_related(product: Product) -> products: list of Product {
      reads: Product
      intent: "every other live product"
    }

    endpoint rate {
      route: POST "/rate"
      input: { product: Product, other: Product, liked: bool }
      output: Review
      steps: rate_product
    }

    step rate_product(product: Product, other: Product, liked: bool) -> review: Review {
      writes: Review
      intent: "store a review of other by product"
    }
`,
	"    component ProductCard(", `    page "/for-me" {
      load: picks from related(product: viewer)
      shows: RateDeck(picks)
      empty: "Nothing like it yet."
    }

    component RateDeck(picks: list of Product): deck {
      shows: picks.name
      says: "Nothing left."
      calls: rate(product: viewer, other: picks, liked)
    }

    component ProductCard(`,
).Replace(viewerApp)

func TestCalls(t *testing.T) {
	app, d := load(map[string]string{"app.philote": callApp})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": callApp}))
	}
	api, web := app.Parts[0], app.Parts[1]
	related := api.Custom[0]
	if related.Method != "GET" || !related.List || related.Output != "Product" || len(related.Flow.Input) != 0 ||
		len(related.Flow.Params) != 1 || related.Flow.Params[0].Name != "product" || !slices.Equal(related.Errors, []string{"not_found"}) {
		t.Errorf("related: %+v", related)
	}
	if r := related.Flow.Steps[0].Result; !r.List || r.Entity != "Product" {
		t.Errorf("find_related returns %+v", r)
	}
	pg := findPage(web, "/for-me")
	if pg.Endpoint != "related" || pg.Resource != "" || pg.Entity != "Product" || pg.ByID || !pg.Viewer {
		t.Errorf("/for-me: %+v", pg)
	}
	call := findComponent(web, "RateDeck").Calls
	var passed []string
	for _, v := range call.Passed {
		passed = append(passed, v.Name)
	}
	if call.Endpoint != "rate" || !slices.Equal(call.Viewer, []string{"product"}) || !slices.Equal(passed, []string{"other", "liked"}) {
		t.Errorf("RateDeck calls %+v, passing %v", call, passed)
	}
	if findPage(web, "/products").Viewer {
		t.Error("/products doesn't need the viewer")
	}
}

func TestCallErrors(t *testing.T) {
	route := `route: GET "/products/[product]/related"`
	runErrorCases(t, callApp, []errorCase{
		{"load from an unknown endpoint", "from related(", "from relatd(", "from »relatd(", "did you mean 'related'"},
		{"load from a POST endpoint", "related(product: viewer)", "rate(product: viewer, other: viewer, liked: viewer)", "from »rate(", "only from GET endpoints"},
		{"load fills an input with something else", "related(product: viewer)", "related(product: picks)", "related(»product: picks)", "with the viewer, like product: viewer"},
		{"load leaves an input out", "related(product: viewer)", "related()", "from »related(", "related needs product"},
		{"load names an unknown input", "related(product: viewer)", "related(prodct: viewer)", "related(»prodct", "did you mean 'product'"},
		{"call a GET endpoint", "calls: rate(product: viewer, other: picks, liked)", "calls: related(product: viewer)", "calls: »related", "is a GET endpoint"},
		{"call leaves a record open", "other: picks, liked)", "other, liked)", "viewer, »other, liked", "fill it with viewer or picks"},
		{"call fills from a stranger", "other: picks, liked)", "other: products, liked)", "other: »products", "isn't the viewer or RateDeck's parameter"},
		{"call fills a scalar with a record", "other: picks, liked)", "other: picks, liked: picks)", "liked: »picks", "liked is a bool"},
		{"call names an input twice", "rate(product: viewer, other", "rate(product: viewer, product: viewer, other", "viewer, »product: viewer", "listed twice"},
		{"GET with a body", "input: { product: Product }", "input: { product: Product, limit: int }", "Product, »limit: int", "a GET endpoint takes no body"},
		{"POST with a param", `route: POST "/rate"`, `route: POST "/rate/[product]"`, `POST »"/rate/[product]"`, "only a GET route takes [params]"},
		{"param that isn't an input", route, `route: GET "/products/[item]/related"`, `GET »"/products/[item]/related"`, "[item] isn't an input"},
		{"param that isn't a record", route + "\n      input: { product: Product }", `route: GET "/related/[product]/[n]"` + "\n      input: { product: Product, n: int }", `GET »"/related/`, "[n] isn't a record"},
		{"route under a relation read", route, `route: GET "/products/[product]/reviews"`, `GET »"/products/[product]/reviews"`, "under the products resource's routes"},
		{"list input", "input: { product: Product }", "input: { product: list of Product }", "{ »product: list of", "a list can't be sent"},
		{"list of a scalar", "-> products: list of Product", "-> products: list of int", "list of »int", "only records come in a list"},
		{"no step returns the list", "-> products: list of Product", "-> products: Product", "output: list of »Product", "no step returns a list of Product"},
	})
}
