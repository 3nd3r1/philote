package spec

import (
	"os"
	"strings"
	"testing"
)

func blogSrc(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../../examples/blog/blog.spec")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestBlogIsClean(t *testing.T) {
	src := blogSrc(t)
	prog, d := Load("blog.spec", src)
	if prog == nil || len(d.List) > 0 {
		t.Fatalf("expected no diagnostics, got:\n%s", Render("blog.spec", src, d.Sorted()))
	}
	if got := FormatFile(prog.File); got != src {
		t.Errorf("fmt does not round-trip:\n%s", got)
	}
}

// Each case edits the blog and expects an error mentioning want at line:col.
func TestErrors(t *testing.T) {
	cases := []struct {
		name, old, new, want, at string
	}{
		{"unknown page key", `  empty: "No posts yet."`, `  empty: "No posts yet."` + "\n  colour: red", "unknown field 'colour' in page", "22:3"},
		{"typo in field path", "post.title, post.published_at, excerpt", "post.titel, post.published_at, excerpt", "Post has no field 'titel'", "31:15"},
		{"markdown as text", "post.body as markdown", "post.body", "markdown can't be shown as plain text", "36:41"},
		{"lookup by non-unique", "Post by slug", "Post by title", "Post.title isn't unique", "25:17"},
		{"route param mismatch", `page "/blog/[slug]"`, `page "/blog/[id]"`, "route param [id] isn't used by data", "24:13"},
		{"list layout on single", "layout: PostView", "layout: list of PostView", "needs a list, but data looks up a single Post", "26:11"},
		{"single layout on list", "layout: list of PostCard", "layout: PostCard", "data returns a list of Post", "20:11"},
		{"order by markdown", "order by published_at desc", "order by body desc", "can't order by Post.body", "19:49"},
		{"compare date with string", "published_at <= now", `published_at <= "yesterday"`, "can't compare Post.published_at (date) with a string", "19:36"},
		{"link passes wrong field", `"/blog/[post.slug]"`, `"/blog/[post.title]"`, "looks up Post by slug, but this link passes Post.title", "32:20"},
		{"link to missing page", `"/blog/[post.slug]"`, `"/posts/[post.slug]"`, "no page matches the link", "32:13"},
		{"missing not_found", "  not_found: 404\n", "", "doesn't say what happens when it's missing", "24:6"},
		{"missing empty", "  empty: \"No posts yet.\"\n", "", "doesn't say what to show when there are no Post records", "18:6"},
		{"unknown type", "published_at: date", "published_at: datetime", "unknown type 'datetime'", "10:17"},
		{"max on non-string", "body: markdown", "body: markdown(max 10)", "(max N) only applies to string", "9:22"},
		{"ambiguous chips", "name: string unique", "name: string unique\n  label: string", "can't tell which field of Tag a chip should show", "37:77"},
		{"unknown renderer", "as chips", "as pills", "unknown renderer 'pills'", "36:77"},
		{"unknown component", "list of PostCard", "list of PostCrad", "did you mean 'PostCard'", "20:19"},
		{"copy too long", `"No posts yet."`, `"` + strings.Repeat("x", 201) + `"`, "UI copy is capped at 200", "21:10"},
		{"single equals", "published_at <= now", "published_at = now", "unexpected '='", "19:33"},
	}
	base := blogSrc(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(base, tc.old) {
				t.Fatalf("test setup: %q not in blog source", tc.old)
			}
			src := strings.Replace(base, tc.old, tc.new, 1)
			prog, d := Load("blog.spec", src)
			if prog != nil {
				t.Fatalf("expected an error, got none")
			}
			for _, x := range d.List {
				if strings.Contains(x.Msg, tc.want) || strings.Contains(x.Hint, tc.want) {
					if x.Pos.String() != tc.at {
						t.Errorf("error at %s, want %s", x.Pos, tc.at)
					}
					return
				}
			}
			t.Errorf("no error containing %q; got:\n%s", tc.want, Render("blog.spec", src, d.Sorted()))
		})
	}
}

func TestRecoversAfterSyntaxError(t *testing.T) {
	src := "app x {\n  target: nextjs\n  style: a\n}\n\nentity A {\n  x: (\n}\n\nentity B {\n  y: strin\n}\n"
	_, d := Load("x.spec", src)
	if len(d.List) < 1 {
		t.Fatal("expected errors")
	}
	// The broken entity A must not hide later problems in the file.
	_, d2 := Load("x.spec", strings.Replace(src, "  x: (\n", "  x: string\n", 1))
	if !strings.Contains(Render("x.spec", src, d2.Sorted()), "did you mean 'string'") {
		t.Errorf("expected a suggestion for 'strin'")
	}
}
