package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A nav body that renders everything its wrapper hands it, and one that
// leaves out the toggle.
const (
	goodNav = "```tsx\nreturn (\n    <nav className=\"flex items-center gap-2 border-b bg-background px-4 py-2\">\n" +
		"        <Link href={home.href}>{home.label}</Link>\n" +
		"        {items.map((item) => (\n" +
		"            <Link key={item.href} href={item.href} aria-current={item.href === active ? \"page\" : undefined}>\n" +
		"                {item.label}\n            </Link>\n        ))}\n" +
		"        {toggle}\n    </nav>\n);\n```"
	navWithoutToggle = "```tsx\nreturn (\n    <nav>\n        <Link href={home.href}>{home.label}</Link>\n" +
		"        {items.map((item) => (\n            <Link key={item.href} href={item.href} aria-current={item.href === active ? \"page\" : undefined}>\n" +
		"                {item.label}\n            </Link>\n        ))}\n    </nav>\n);\n```"
)

// A nav line goes through the same loop as a body: problems go back to the
// model, a passing nav is pinned, and changing an item rewrites the nav and
// nothing else.
func TestNavFromSpec(t *testing.T) {
	spec := copyLibrary(t)
	setNav(t, spec, `nav: side, "/authors", "/books" as "Shelf"`)
	out := filepath.Join(filepath.Dir(spec), "build")
	linkNodeModules(t, "testdata/library/want/build/web/package.json", filepath.Join(out, "web"))
	r := useReplies(t, map[string][]string{
		"NavBody":        {navWithoutToggle, goodNav},
		"AuthorCardBody": {goodAuthorCard},
		"AuthorViewBody": {goodAuthorView},
		"BookShelfBody":  {goodBookShelf},
	})
	runOK(t, "build", "--out", out, spec)
	if n := len(r.prompts["NavBody"]); n != 2 || !strings.Contains(r.prompts["NavBody"][1], "{toggle} exactly once") {
		t.Fatalf("the nav was asked %d times, want 2 with the toggle problem in the retry", n)
	}
	if p := r.prompts["NavBody"][0]; !strings.Contains(p, "a column on the left") || !strings.Contains(p, "/books: Shelf") {
		t.Errorf("the nav prompt doesn't hold the place and items:\n%s", p)
	}
	if lock := readFile(t, filepath.Join(filepath.Dir(spec), "philote.lock")); !strings.Contains(lock, "nav web/nav\n") {
		t.Errorf("want the nav pinned as web/nav:\n%s", lock)
	}
	navTS := readFile(t, filepath.Join(out, "web/src/lib/nav.ts"))
	if !strings.Contains(navTS, `= "side";`) || !strings.Contains(navTS, `{ href: "/books", label: "Shelf" }`) {
		t.Errorf("nav.ts doesn't hold the spec's nav:\n%s", navTS)
	}

	setNav(t, spec, `nav: side, "/authors", "/books" as "Books"`)
	r = useReplies(t, map[string][]string{"NavBody": {goodNav}})
	runOK(t, "build", "--out", out, spec)
	if len(r.prompts) != 1 || len(r.prompts["NavBody"]) != 1 || !strings.Contains(r.prompts["NavBody"][0], "An earlier version") {
		t.Errorf("a changed item should ask for the nav only, starting from the old one; got %v", r.prompts)
	}
}

// setNav puts a nav line in the library spec, or replaces it.
func setNav(t *testing.T, spec, nav string) {
	t.Helper()
	data := readFile(t, spec)
	if i := strings.Index(data, "    nav: "); i >= 0 {
		end := strings.Index(data[i:], "\n")
		data = data[:i] + data[i+end+1:]
	}
	data = strings.Replace(data, "    uses: api\n", "    uses: api\n    "+nav+"\n", 1)
	if err := os.WriteFile(spec, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
