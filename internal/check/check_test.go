package check

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/3nd3r1/philote/internal/diag"
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/syntax"
)

const base = `app shop {
  backend api: hono {
    resource products: Product
  }
}

enum Status { draft, live }

entity Product {
  name: string(max 120)
  slug: slug from name
  price: int(min 0)
  status: Status
}

entity Review {
  stars: int(min 1, max 5)
}

Product has many reviews: Review
`

func load(files map[string]string) (*ir.App, *diag.List) {
	fsys := fstest.MapFS{}
	for name, src := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(src)}
	}
	d := &diag.List{}
	src := syntax.Load(fsys, "app.philote", d)
	if d.HasErrors() {
		return nil, d
	}
	return Run(src, Targets{"hono": ir.Backend}, d), d
}

func TestBaseIsClean(t *testing.T) {
	app, d := load(map[string]string{"app.philote": base})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": base}))
	}
	review := app.Entity("Review")
	if len(review.Relations) != 1 || review.Relations[0].FK != "product_id" {
		t.Errorf("Review should get a product relation with key product_id, got %+v", review.Relations)
	}
}

// Each case replaces old with new in base. at is the new text with » where
// the error must point; want must appear in the message or hint.
func TestErrors(t *testing.T) {
	cases := []struct{ name, old, new, at, want string }{
		{"unknown type", "price: int(min 0)", "price: money", "price: »money", "unknown type 'money'"},
		{"suggests a close name", "status: Status", "status: Stauts", "status: »Stauts", "did you mean 'Status'"},
		{"entity in a field", "status: Status", "status: Review", "status: »Review", "use a relation"},
		{"reserved id", "price: int(min 0)", "id: string", "»id: string", "already defined by the automatic id"},
		{"limit on an enum", "status: Status", "status: Status(max 3)", "Status(»max", "only applies to string and int"},
		{"min above max", "min 1, max 5", "min 6, max 5", "(»min 6", "min is larger than max"},
		{"slug from an int", "slug from name", "slug from price", "from »price", "only be made from a string field"},
		{"duplicate enum value", "draft, live", "draft, live, draft", "live, »draft", "listed twice"},
		{"relation both ways", "Product has many reviews: Review", "Product has many reviews: Review\nReview belongs to product: Product",
			"belongs to »product", "Review.product is already defined"},
		{"reverse name clash", "stars: int(min 1, max 5)", "stars: int(min 1, max 5)\n  product: string",
			"reviews: »Review", "(back <name>)"},
		{"unknown target", "backend api: hono", "backend api: express", "api: »express", "unknown target 'express'"},
		{"wrong kind of target", "  backend api: hono {", "  frontend web: hono {}\n  backend api: hono {", "web: »hono", "hono can't build a frontend"},
		{"unknown operation", "resource products: Product", "resource products: Product { only: list, fetch }", "list, »fetch", "unknown operation 'fetch'"},
		{"unknown resource entity", "resource products: Product", "resource products: Prodct", "products: »Prodct", "did you mean 'Product'"},
		{"resource outside a part", "enum Status", "resource extra: Product\nenum Status", "resource »extra", "belongs inside a part"},
		{"missing colon", "name: string(max 120)", "name string(max 120)", "name »string", "expected ':'"},
		{"unknown statement", "enum Status", "table Foo\nenum Status", "»table Foo", "unknown statement 'table'"},
		{"missing body file", "backend api: hono {\n    resource products: Product\n  }", "backend api: hono", "backend »api", "api.philote doesn't exist"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(base, tc.old) {
				t.Fatalf("test setup: %q is not in base", tc.old)
			}
			src := strings.Replace(base, tc.old, tc.new, 1)
			_, d := load(map[string]string{"app.philote": src})
			want := posOf(t, src, tc.at)
			for _, x := range d.Items {
				if strings.Contains(x.Msg+" "+x.Hint, tc.want) {
					if x.Pos != want {
						t.Errorf("error at %s, want %s", x.Pos, want)
					}
					return
				}
			}
			t.Errorf("no diagnostic containing %q; got:\n%s", tc.want, diag.Render(d.Items, map[string]string{"app.philote": src}))
		})
	}
}

// posOf finds at (with » removed) in src and returns where » points.
func posOf(t *testing.T, src, at string) diag.Pos {
	t.Helper()
	before, _, _ := strings.Cut(at, "»")
	i := strings.Index(src, strings.Replace(at, "»", "", 1))
	if i < 0 {
		t.Fatalf("test setup: %q not found in source", at)
	}
	i += len(before)
	line := strings.Count(src[:i], "\n") + 1
	col := i - strings.LastIndex(src[:i], "\n")
	return diag.Pos{File: "app.philote", Line: line, Col: col}
}

func TestReportsErrorsInSeveralEntities(t *testing.T) {
	src := strings.Replace(base, "price: int(min 0)", "price int", 1)
	src = strings.Replace(src, "stars: int(min 1, max 5)", "stars int", 1)
	_, d := load(map[string]string{"app.philote": src})
	if len(d.Items) != 2 {
		t.Errorf("want one error in each entity, got:\n%s", diag.Render(d.Items, map[string]string{"app.philote": src}))
	}
}

func TestBodyFromFile(t *testing.T) {
	root := strings.Replace(base, "backend api: hono {\n    resource products: Product\n  }", "backend api: hono", 1)
	app, d := load(map[string]string{"app.philote": root, "api.philote": "resource products: Product\n"})
	if d.HasErrors() || len(app.Parts[0].Resources) != 1 {
		t.Fatalf("want api's body loaded from api.philote, got %+v", d.Items)
	}
}
