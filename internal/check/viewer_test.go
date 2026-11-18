package check

import (
	"strings"
	"testing"

	"github.com/3nd3r1/philote/internal/diag"
)

// viewerApp is the views app where a visitor picks the product they are on
// /pick and sees it on /me.
var viewerApp = strings.NewReplacer(
	"    uses: api\n", "    uses: api\n    viewer: Product\n",
	"    component ProductCard(", `    page "/pick" {
      load: products
      shows: ProductPicker(products)
      empty: "No products yet."
    }

    page "/me" {
      shows: ProductView(viewer)
    }

    component ProductPicker(products: list of Product) {
      shows: products.name
      picks: viewer then "/me"
    }

    component ProductCard(`,
).Replace(views)

func TestViewer(t *testing.T) {
	app, d := load(map[string]string{"app.philote": viewerApp})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": viewerApp}))
	}
	web := app.Parts[1]
	if v := web.Viewer; v == nil || v.Entity != "Product" || v.Resource != "products" || v.Picker != "/pick" || v.Pos.Line != 8 {
		t.Fatalf("viewer: %+v", v)
	}
	if picker := findComponent(web, "ProductPicker"); picker.Picks != "/me" {
		t.Errorf("ProductPicker picks then %q", picker.Picks)
	}
	me := findPage(web, "/me")
	if len(me.Sections) != 1 || !me.Sections[0].Viewer || !me.Sections[0].Arg || me.Load != "" {
		t.Errorf("/me: %+v", me.Sections[0])
	}
}

func TestViewerErrors(t *testing.T) {
	runErrorCases(t, viewerApp, []errorCase{
		{"unknown entity", "viewer: Product", "viewer: Prodct", "viewer: »Prodct", "did you mean 'Product'"},
		{"no resource gets it", "viewer: Product", "viewer: Review", "viewer: »Review", "gets a Review by id"},
		{"two viewers", "    viewer: Product\n", "    viewer: Product\n    viewer: Product\n", "Product\n    »viewer: Product", "already has a viewer"},
		{"no viewer line", "    viewer: Product\n", "", "ProductView(»viewer)", "web has no viewer"},
		{"wrong entity", "shows: ProductView(viewer)", "shows: ReviewCard(viewer)", "ReviewCard(»viewer)", "the viewer is a Product"},
		{"a list of the viewer", "shows: ProductView(viewer)", "shows: list of ProductCard(viewer)", "list of »ProductCard(viewer)", "the viewer is one record"},
		{"picks something else", "picks: viewer then", "picks: owner then", "picks: »owner", "only the viewer"},
		{"picks one record", "      links_to: \"/products/[product.id]\"\n", "      links_to: \"/products/[product.id]\"\n      picks: viewer then \"/me\"\n",
			"component »ProductCard(", "takes the list to pick from"},
		{"picks then a missing page", `picks: viewer then "/me"`, `picks: viewer then "/mee"`, `then »"/mee"`, "did you mean '/me'"},
		{"picks then a page by id", `picks: viewer then "/me"`, `picks: viewer then "/products/[id]"`, `then »"/products/[id]"`, "needs an id"},
		{"no picker", "      picks: viewer then \"/me\"\n", "", "»viewer: Product", "no page picks it"},
		{"two pickers", "      shows: list of ProductCard(products)\n", "      shows: ProductPicker(products)\n", "»viewer: Product", "both pick the viewer"},
		{"picker shows the viewer", "      shows: ProductPicker(products)\n", "      shows: ProductPicker(products), ProductView(viewer)\n", "»viewer: Product", "can't need it too"},
		{"load named viewer", "load: product from products by id", "load: viewer from products by id", "load: »viewer", "name what the page loads something else"},
	})
}
