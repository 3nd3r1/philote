package check

import (
	"slices"
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

    page "/products" {
      data: products
      layout: list of ProductCard
      empty: "No products yet."
    }

    page "/products/[id]" {
      data: products by id
      layout: ProductView
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
	if len(web.Pages) != 2 || !web.Pages[1].ByID || web.Pages[0].Empty != "No products yet." {
		t.Errorf("pages: %+v", web.Pages)
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
		{"page without data", "      data: products\n      layout: list", "      layout: list", `»page "/products" {`, "needs data"},
		{"unknown resource", "data: products\n", "data: prodcts\n", "data: »prodcts", "did you mean 'products'"},
		{"load by another field", "data: products by id", "data: products by slug", "by »slug", "by id only"},
		{"list without empty", `      empty: "No products yet."` + "\n", "", `»page "/products" {`, "say what shows when it's empty"},
		{"by id without not_found", `      not_found: "That product doesn't exist."` + "\n", "", `»page "/products/[id]"`, "when the record doesn't exist"},
		{"by id without [id]", `page "/products/[id]"`, `page "/product"`, `page »"/product"`, "needs [id] in its route"},
		{"list with [id]", `page "/products" {`, `page "/products/[id]/all" {`, "data: »products\n", "load one record with `data: products by id`"},
		{"list layout without list of", "layout: list of ProductCard", "layout: ProductCard", "layout: »ProductCard", "use `layout: list of ProductCard`"},
		{"wrong entity", "layout: ProductView", "layout: ReviewCard", "layout: »ReviewCard", "ReviewCard shows a Review"},
		{"unknown component", "layout: ProductView", "layout: ProductViw", "layout: »ProductViw", "did you mean 'ProductView'"},
		{"home page", `page "/products" {`, `page "/" {`, `page »"/" {`, "home page"},
		{"bad route", `page "/products" {`, `page "/Products" {`, `page »"/Products" {`, "lowercase segments"},
		{"lowercase component", "component ReviewCard(", "component reviewCard(", "component »reviewCard", "capital letter"},
		{"two params", "ReviewCard(review: Review)", "ReviewCard(review: Review, other: Review)", "component »ReviewCard", "takes one record"},
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
		{"link by a missing field", "/products/[product.id]", "/products/[product.sku]", `links_to: »"/products/[product.sku]"`, "no field 'sku'"},
	})
}
