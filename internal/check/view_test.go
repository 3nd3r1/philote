package check

import (
	"slices"
	"strings"
	"testing"

	"github.com/3nd3r1/philote/internal/diag"
)

const views = `app shop {
  backend api: hono {
    resource products: Product
    resource reviews: Review { only: list }
  }
  frontend web: nextjs {
    uses: api

    page "/" {
      load: products
      shows: Hero, ProductGrid(products)
      empty: "Nothing for sale yet."
    }

    page "/products" {
      load: products
      shows: list of ProductCard(products)
      empty: "No products yet."
    }

    page "/products/[id]" {
      load: product from products by id
      shows: ProductView(product)
      not_found: "That product doesn't exist."
    }

    component ProductCard(product: Product) {
      shows: product.name, product.price
      links_to: "/products/[product.id]"
    }

    component ProductView(product: Product) {
      shows: product.name, product.status, product.reviews as chips
    }

    component ReviewCard(review: Review) {
      shows: review.stars, review.product
    }

    component Hero {
      says: "Good things.", "Made slowly."
      links_to: "/products" as "Shop"
    }

    component ProductGrid(products: list of Product) {
      shows: products.name, products.price
    }
  }
}

enum Status { draft, live }

entity Product {
  display: name
  name: string(max 120)
  slug: slug from name
  price: int(min 0)
  status: Status
  body: markdown
}

entity Review {
  display: title
  title: string(max 80)
  stars: int(min 1, max 5)
}

Product has many reviews: Review
`

func TestViewsResolve(t *testing.T) {
	app, d := load(map[string]string{"app.philote": views})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": views}))
	}
	web := app.Parts[1]
	home, list, one := web.Pages[0], web.Pages[1], web.Pages[2]
	if home.Load != "products" || len(home.Sections) != 2 || home.Sections[0].Arg || !home.Sections[1].Arg || home.Sections[1].List {
		t.Errorf("home: %+v", home)
	}
	if list.Empty != "No products yet." || !list.Sections[0].List || !one.ByID || one.Load != "product" || one.Resource != "products" {
		t.Errorf("pages: %+v, %+v", list, one)
	}
	hero, grid := web.Components[3], web.Components[4]
	if hero.Param != "" || !slices.Equal(hero.Says, []string{"Good things.", "Made slowly."}) || hero.LinksTo != "/products" || hero.LinkLabel != "Shop" {
		t.Errorf("Hero: %+v", hero)
	}
	if !grid.List || grid.Param != "products" || len(grid.Shows) != 2 {
		t.Errorf("ProductGrid: %+v", grid)
	}
	card, view, review := web.Components[0], web.Components[1], web.Components[2]
	if card.LinksTo != "/products/[id]" || !slices.Equal(card.LinkBy, []string{"id"}) {
		t.Errorf("ProductCard links to %q by %v", card.LinksTo, card.LinkBy)
	}
	// Both relation reads go through products: its get for a review's
	// product, its relation route for a product's reviews.
	if r := view.Shows[2]; r.Relation == nil || !r.Relation.Many || r.Resource != "products" || r.Display != "title" {
		t.Errorf("product.reviews: %+v", r)
	}
	if r := review.Shows[1]; r.Relation == nil || r.Relation.Many || r.Resource != "products" || r.Display != "name" {
		t.Errorf("review.product: %+v", r)
	}
}

func TestViewErrors(t *testing.T) {
	runErrorCases(t, views, []errorCase{
		{"page shows nothing", "      shows: Hero, ProductGrid(products)\n", "", `»page "/" {`, "doesn't show anything"},
		{"load shown by nothing", "shows: Hero, ProductGrid(products)", "shows: Hero", "load: »products", "shows nothing with it"},
		{"unknown resource", "from products by id", "from prodcts by id", "from »prodcts", "did you mean 'products'"},
		{"load by another field", "from products by id", "from products by slug", "by »slug", "by id only"},
		{"list without empty", `      empty: "Nothing for sale yet."` + "\n", "", `»page "/" {`, "say what shows when it's empty"},
		{"by id without not_found", `      not_found: "That product doesn't exist."` + "\n", "", `»page "/products/[id]"`, "when the record doesn't exist"},
		{"by id without [id]", `page "/products/[id]"`, `page "/product"`, `page »"/product"`, "needs [id] in its route"},
		{"list with [id]", `page "/products" {`, `page "/products/[id]/all" {`, "      load: »products\n      shows: list", "load one record with `load: products by id`"},
		{"[id] without load", "      load: product from products by id\n", "", `page »"/products/[id]"`, "loads nothing"},
		{"one record per list item", "shows: list of ProductCard(products)", "shows: ProductCard(products)", "ProductCard(»products)", "use list of ProductCard(products)"},
		{"list of a list", "shows: Hero, ProductGrid(products)", "shows: Hero, list of ProductGrid(products)", "list of »ProductGrid", "already takes a list"},
		{"list of one record", "shows: ProductView(product)", "shows: list of ProductView(product)", "list of »ProductView", "is one record"},
		{"wrong entity", "shows: ProductView(product)", "shows: ReviewCard(product)", "ReviewCard(»product)", "ReviewCard shows Review records"},
		{"unknown component", "shows: ProductView(product)", "shows: ProductViw(product)", "shows: »ProductViw", "did you mean 'ProductView'"},
		{"missing argument", "shows: ProductView(product)", "shows: ProductView", "shows: »ProductView", "pass it what the page loads"},
		{"unknown argument", "shows: ProductView(product)", "shows: ProductView(item)", "ProductView(»item)", "loads nothing named item"},
		{"argument for no record", "shows: Hero, ProductGrid", "shows: Hero(products), ProductGrid", "Hero(»products)", "takes no record"},
		{"bad route", `page "/products" {`, `page "/Products" {`, `page »"/Products" {`, "lowercase segments"},
		{"recordless shows", "component Hero {\n", "component Hero {\n      shows: hero.name\n", "shows: »hero.name", "nothing to show"},
		{"recordless says nothing", `      says: "Good things.", "Made slowly."` + "\n", "", "component »Hero", "says nothing"},
		{"link without a label", `links_to: "/products" as "Shop"`, `links_to: "/products"`, `links_to: »"/products"` + "\n", "needs a button text"},
		{"says too long", `"Made slowly."`, `"` + strings.Repeat("a", 201) + `"`, `"Good things.", »"aaa`, "capped at 200"},
		{"says empty", `"Made slowly."`, `""`, `"Good things.", »""`, "this text is empty"},
		{"list with a link", "      shows: products.name, products.price\n", "      shows: products.name, products.price\n      links_to: \"/products\"\n", "ProductGrid(»products: list", "can't have links_to or actions"},
		{"lowercase component", "component ReviewCard(", "component reviewCard(", "component »reviewCard", "capital letter"},
		{"two params", "ReviewCard(review: Review)", "ReviewCard(review: Review, other: Review)", "Review, »other: Review", "one record or one list"},
		{"unknown member", "product.name, product.price", "product.nme, product.price", "product.»nme", "did you mean 'name'"},
		{"wrong parameter", "shows: review.stars", "shows: rev.stars", "shows: »rev.stars", "isn't ReviewCard's parameter"},
		{"deep path", "review.product\n", "review.product.name\n", "review.stars, »review.product.name", "one level below"},
		{"markdown", "product.name, product.price", "product.name, product.body", "product.»body", "markdown fields can't be shown yet"},
		{"list without chips", "product.reviews as chips", "product.reviews", "product.»reviews", "show it `as chips`"},
		{"chips on a field", "product.name, product.price", "product.name as chips, product.price", "product.name as »chips", "only a has-many relation"},
		{"unknown format", "product.reviews as chips", "product.reviews as tags", "product.reviews as »tags", "unknown format 'tags'"},
		{"shown twice", "product.name, product.price", "product.name, product.name", "product.name, product.»name", "shown twice"},
		{"no display field", "  display: name\n", "", "review.»product", "add `display: <field>` to Product"},
		{"no resource to load with", "resource products: Product\n", "resource products: Product { only: list }\n", "product.»reviews as chips", "add `resource <name>: Product` to api"},
		{"link to a missing page", "/products/[product.id]", "/product/[product.id]", `links_to: »"/product/[product.id]"`, "no page at /product/[id]"},
		{"look too long", "      links_to: \"/products/[product.id]\"\n", "      links_to: \"/products/[product.id]\"\n      look: \"" + strings.Repeat("a", 201) + "\"\n",
			"      look: »\"aaa", "the cap is 200"},
		{"link by a missing field", "/products/[product.id]", "/products/[product.sku]", `links_to: »"/products/[product.sku]"`, "no field 'sku'"},
		{"app look too long", "    uses: api\n", "    uses: api\n    look: \"" + strings.Repeat("a", 201) + "\"\n", "    look: »\"aaa", "the cap is 200"},
		{"app look empty", "    uses: api\n", "    uses: api\n    look: \"\"\n", `look: »""`, "the look is empty"},
		{"two app looks", "    uses: api\n", "    uses: api\n    look: \"warm\"\n    look: \"cold\"\n", `»look: "cold"`, "already has a look"},
		{"look in a backend", "    resource products: Product\n", "    look: \"warm\"\n    resource products: Product\n", `»look: "warm"`, "can't go inside a backend"},
		{"nav place", "    uses: api\n", "    uses: api\n    nav: middle, \"/products\"\n", "nav: »middle", "top, bottom or side, not middle"},
		{"nav without pages", "    uses: api\n", "    uses: api\n    nav: top\n", "»nav: top", "the nav has no pages"},
		{"nav to a missing page", "    uses: api\n", "    uses: api\n    nav: top, \"/prodcts\"\n", `top, »"/prodcts"`, "did you mean '/products'"},
		{"nav to a page by id", "    uses: api\n", "    uses: api\n    nav: top, \"/products/[id]\"\n", `top, »"/products/[id]"`, "needs an id"},
		{"nav item twice", "    uses: api\n", "    uses: api\n    nav: top, \"/products\", \"/products\"\n", `"/products", »"/products"`, "already in the nav"},
		{"nav label empty", "    uses: api\n", "    uses: api\n    nav: top, \"/products\" as \"\"\n", `as »""`, "use the page's title"},
		{"nav label too long", "    uses: api\n", "    uses: api\n    nav: top, \"/products\" as \"" + strings.Repeat("a", 25) + "\"\n", `as »"aaa`, "capped at 24"},
		{"two navs", "    uses: api\n", "    uses: api\n    nav: top, \"/\"\n    nav: side, \"/\"\n", "»nav: side", "already has a nav"},
		{"nav in a backend", "    resource products: Product\n", "    nav: top, \"/\"\n    resource products: Product\n", "»nav: top", "can't go inside a backend"},
	})
}

// A frontend's look line reaches the IR for the theme's prompt.
func TestAppLook(t *testing.T) {
	src := strings.Replace(views, "    uses: api\n", "    uses: api\n    look: \"warm and playful\"\n", 1)
	app, d := load(map[string]string{"app.philote": src})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": src}))
	}
	if l := app.Parts[1].Look; l == nil || l.Text != "warm and playful" || l.Pos.Line != 8 {
		t.Errorf("look: %+v", l)
	}
}

// A nav line reaches the IR in order, with a label only where the spec
// gives one.
func TestNav(t *testing.T) {
	src := strings.Replace(views, "    uses: api\n", "    uses: api\n    nav: bottom, \"/\" as \"Home\",\n      \"/products\"\n", 1)
	app, d := load(map[string]string{"app.philote": src})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": src}))
	}
	nav := app.Parts[1].Nav
	if nav == nil || nav.Place != "bottom" || len(nav.Items) != 2 || nav.Pos.Line != 8 {
		t.Fatalf("nav: %+v", nav)
	}
	if a, b := nav.Items[0], nav.Items[1]; a.Route != "/" || a.Label != "Home" || b.Route != "/products" || b.Label != "" {
		t.Errorf("items: %+v, %+v", a, b)
	}
}

// A shown relation links to its entity's page when exactly one page loads
// that entity by id, and the component isn't a link itself (D-024).
func TestRelationLinks(t *testing.T) {
	secondPage := `    component ProductCard(`
	reviewPage := `    page "/reviews/[id]" {
      load: review from reviews by id
      shows: ReviewCard(review)
      not_found: "That review doesn't exist."
    }

    component ProductCard(`
	cases := []struct {
		name, old, new string
		want, reviews  string // where review.product and product.reviews link
	}{
		{"one page", "", "", "/products/[id]", ""},
		{"two pages", secondPage, "    page \"/items/[id]\" {\n      load: product from products by id\n      shows: ProductView(product)\n      not_found: \"Gone.\"\n    }\n\n" + secondPage, "", ""},
		{"linked component", "      shows: review.stars, review.product\n", "      shows: review.stars, review.product\n      links_to: \"/reviews/[review.id]\"\n", "", "/reviews/[id]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := strings.Replace(views, tc.old, tc.new, 1)
			if tc.name == "linked component" {
				src = strings.Replace(src, "{ only: list }", "{ only: list, get }", 1)
				src = strings.Replace(src, secondPage, reviewPage, 1)
			}
			app, d := load(map[string]string{"app.philote": src})
			if len(d.Items) > 0 {
				t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": src}))
			}
			review := findComponent(app.Parts[1], "ReviewCard")
			if got := review.Shows[1].Href; got != tc.want {
				t.Errorf("review.product links to %q, want %q", got, tc.want)
			}
			if got := findComponent(app.Parts[1], "ProductView").Shows[2].Href; got != tc.reviews {
				t.Errorf("product.reviews links to %q, want %q", got, tc.reviews)
			}
		})
	}
}
