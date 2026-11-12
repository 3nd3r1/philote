package check

import (
	"slices"
	"testing"

	"github.com/3nd3r1/philote/internal/diag"
	"github.com/3nd3r1/philote/internal/ir"
)

const forms = `app shop {
  backend api: hono {
    resource products: Product
    resource reviews: Review { only: list, get, create, delete }

    endpoint feature {
      route: POST "/feature"
      input: { product: Product, live: bool, rank: int(min 1, max 9) }
      output: Product
      steps: set_featured
    }

    step set_featured(product: Product, live: bool, rank: int(min 1, max 9)) -> featured: Product {
      intent: "mark the product featured at that rank"
    }
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

    page "/products/new" {
      form: products.create
      then: "/products/[id]"
    }

    page "/products/[id]/edit" {
      form: products.update by id
      then: "/products/[id]"
      not_found: "That product doesn't exist."
    }

    page "/reviews/new" {
      form: reviews.create
      then: "/products"
    }

    page "/feature" {
      form: feature
      then: "/products/[id]"
    }

    component ProductCard(product: Product) {
      shows: product.name
      links_to: "/products/[product.id]"
      actions: delete as "Remove"
    }

    component ProductView(product: Product) {
      shows: product.name, product.status
      actions: open "/products/[product.id]/edit", delete then "/products"
    }
  }
}

enum Status { draft, live }

entity Product {
  display: name
  name: string(max 120)
  status: Status
}

entity Review {
  stars: int(min 1, max 5)
}

Product has many reviews: Review
`

func TestFormsResolve(t *testing.T) {
	app, d := load(map[string]string{"app.philote": forms})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": forms}))
	}
	web := app.Parts[1]
	create, edit, review, feature := web.Pages[2], web.Pages[3], web.Pages[4], web.Pages[5]
	if create.Form != "products.create" || create.ByID || create.Then != "/products/[id]" {
		t.Errorf("create: %+v", create)
	}
	if edit.Form != "products.update" || !edit.ByID || edit.Entity != "Product" || edit.NotFound == "" {
		t.Errorf("edit: %+v", edit)
	}
	if review.Form != "reviews.create" || review.Then != "/products" {
		t.Errorf("review: %+v", review)
	}
	if feature.Form != "feature" || feature.Resource != "" {
		t.Errorf("feature: %+v", feature)
	}
	if a := web.Components[0].Actions[0]; a.Kind != ir.Remove || a.Label != "Remove" || a.Then != "" {
		t.Errorf("card delete: %+v", a)
	}
	open, del := web.Components[1].Actions[0], web.Components[1].Actions[1]
	if open.Kind != ir.Open || open.Route != "/products/[id]/edit" || !slices.Equal(open.RouteBy, []string{"id"}) {
		t.Errorf("open: %+v", open)
	}
	if del.Kind != ir.Remove || del.Endpoint != "products.delete" || del.Then != "/products" {
		t.Errorf("delete: %+v", del)
	}
}

func TestFormErrors(t *testing.T) {
	runErrorCases(t, forms, []errorCase{
		{"form with data", "      form: products.create\n", "      form: products.create\n      data: products\n", "form: »products.create", "not data"},
		{"unknown resource", "form: products.create", "form: prodcts.create", "form: »prodcts", "did you mean 'products'"},
		{"form on a read", "form: products.create", "form: products.list", "form: products.»list", "creates or updates"},
		{"op not served", "form: reviews.create", "form: reviews.update by id", "form: reviews.»update", "doesn't serve update"},
		{"edit without by id", "form: products.update by id", "form: products.update", "form: »products.update", "use `form: products.update by id`"},
		{"edit without [id]", `page "/products/[id]/edit"`, `page "/products/edit"`, `page »"/products/edit"`, "needs [id] in its route"},
		{"create with [id]", `page "/products/new"`, `page "/products/[id]/copy"`, `page »"/products/[id]/copy"`, "can't have [id]"},
		{"edit without not_found", "      then: \"/products/[id]\"\n      not_found: \"That product doesn't exist.\"\n", "      then: \"/products/[id]\"\n", `»page "/products/[id]/edit"`, "when the record doesn't exist"},
		{"form without then", "      form: reviews.create\n      then: \"/products\"\n", "      form: reviews.create\n", `»page "/reviews/new"`, "where to go after saving"},
		{"then to a missing page", `then: "/products"`, `then: "/prodcts"`, `then: »"/prodcts"`, "no page at /prodcts"},
		{"then [id] of another entity", "reviews.create\n      then: \"/products\"", "reviews.create\n      then: \"/products/[id]\"", "reviews.create\n      then: »\"/products/[id]\"", "doesn't load a Review"},
		{"form on a resource without op", "form: feature", "form: products", "form: »products\n", "products.create or products.update"},
		{"then on a data page", "      empty: \"No products yet.\"\n", "      empty: \"No products yet.\"\n      then: \"/products\"\n", `then: »"/products"`, "only a form page has then"},
		{"unknown endpoint", "form: feature", "form: featur", "form: »featur", "did you mean 'feature'"},
		{"no list for a picked record", "resource products: Product\n", "resource products: Product { only: get, create, update, delete }\n", "form: »reviews", "no resource that lists Product"},
		{"unknown action", "delete then", "delet then", "»delet then", "did you mean 'delete'"},
		{"delete twice", `delete then "/products"`, `delete then "/products", delete`, `"/products", »delete`, "already has a delete"},
		{"no resource deletes", "resource products: Product\n", "resource products: Product { only: list, get, create, update }\n", "»delete as", "no resource that deletes Product"},
		{"open with then", `open "/products/[product.id]/edit"`, `open "/products/[product.id]/edit" then "/products"`, `then »"/products"`, "then doesn't apply"},
		{"open a missing page", `open "/products/[product.id]/edit"`, `open "/products/[product.id]/change"`, `open »"/products`, "no page at /products/[id]/change"},
	})
}
