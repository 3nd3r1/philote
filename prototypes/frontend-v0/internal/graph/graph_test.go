package graph

import (
	"os"
	"sort"
	"strings"
	"testing"

	"specc/internal/spec"
)

var cfg = Config{CompilerVersion: "test", PromptVersion: "1", Model: "m"}

func keys(t *testing.T, src string) map[string]string {
	t.Helper()
	prog, d := spec.Load("blog.spec", src)
	if prog == nil {
		t.Fatalf("load failed:\n%s", spec.Render("blog.spec", src, d.Sorted()))
	}
	out := map[string]string{}
	for _, n := range Build(prog, cfg) {
		out[n.ID] = n.Key
	}
	return out
}

func changed(a, b map[string]string) []string {
	var out []string
	for id, k := range b {
		if a[id] != k {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// These pin down experiment step 4 at the cache-key level: an edit rebuilds
// exactly the nodes listed, and nothing else.
func TestLocality(t *testing.T) {
	b, err := os.ReadFile("../../examples/blog/blog.spec")
	if err != nil {
		t.Fatal(err)
	}
	base := string(b)
	cases := []struct {
		name     string
		old, new string
		want     []string
	}{
		{"whitespace and comments", "app blog {", "// my blog\n\n\napp   blog {", nil},
		{"PostCard body", "excerpt(post.body, 160)", "excerpt(post.body, 200)", []string{"component:PostCard"}},
		{"PostCard drops a field", "post.title, post.published_at, excerpt", "post.title, excerpt", []string{"component:PostCard"}},
		{"PostCard param rename",
			"PostCard(post: Post) {\n  shows: post.title, post.published_at, excerpt(post.body, 160)\n  links_to: \"/blog/[post.slug]\"",
			"PostCard(p: Post) {\n  shows: p.title, p.published_at, excerpt(p.body, 160)\n  links_to: \"/blog/[p.slug]\"",
			[]string{"component:PostCard", "page:/blog"}},
		{"empty copy", "No posts yet.", "Nothing here.", []string{"page:/blog"}},
		{"Tag gains a field", "name: string unique", "name: string unique\n  color: int", []string{"component:PostView", "entity:Post", "entity:Tag"}},
		{"Post gains a field", "tags: [Tag]", "tags: [Tag]\n  draft: bool", []string{
			"component:PostCard", "component:PostView", "entity:Post", "page:/blog", "page:/blog/[slug]"}},
	}
	k0 := keys(t, base)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(base, tc.old) {
				t.Fatalf("test setup: %q not in source", tc.old)
			}
			got := changed(k0, keys(t, strings.Replace(base, tc.old, tc.new, 1)))
			if strings.Join(got, " ") != strings.Join(tc.want, " ") {
				t.Errorf("rebuilt %v, want %v", got, tc.want)
			}
		})
	}
}
