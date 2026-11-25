package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A nav line renders a templated top, bottom or side nav with zero model
// calls for the nav and no web/nav pin in philote.lock.
func TestNavFromSpec(t *testing.T) {
	spec := copyLibrary(t)
	setNav(t, spec, `nav: side, "/authors", "/books" as "Shelf"`)
	out := filepath.Join(filepath.Dir(spec), "build")
	linkNodeModules(t, "testdata/library/want/build/web/package.json", filepath.Join(out, "web"))
	r := useReplies(t, map[string][]string{
		"AuthorCardBody": {goodAuthorCard},
		"AuthorViewBody": {goodAuthorView},
		"BookShelfBody":  {goodBookShelf},
	})
	runOK(t, "build", "--out", out, spec)
	if n := len(r.prompts["NavBody"]); n != 0 {
		t.Fatalf("the nav asked the model %d times, want 0", n)
	}
	if lock := readFile(t, filepath.Join(filepath.Dir(spec), "philote.lock")); strings.Contains(lock, "web/nav") {
		t.Errorf("templated nav should not be pinned in philote.lock:\n%s", lock)
	}
	navTS := readFile(t, filepath.Join(out, "web/src/lib/nav.ts"))
	if !strings.Contains(navTS, `= "side";`) || !strings.Contains(navTS, `{ href: "/books", label: "Shelf" }`) {
		t.Errorf("nav.ts doesn't hold the spec's nav:\n%s", navTS)
	}
	bar := readFile(t, filepath.Join(out, "web/src/components/common/layouts/main-navbar.tsx"))
	if !strings.Contains(bar, "<aside") {
		t.Errorf("side nav should render an <aside>:\n%s", bar)
	}

	setNav(t, spec, `nav: bottom, "/authors", "/books" as "Books"`)
	r = useReplies(t, nil)
	runOK(t, "build", "--frozen", "--out", out, spec)
	if len(r.prompts) != 0 {
		t.Errorf("changing nav should make 0 model calls; got %v", r.prompts)
	}
	bar = readFile(t, filepath.Join(out, "web/src/components/common/layouts/main-navbar.tsx"))
	if !strings.Contains(bar, "border-t") {
		t.Errorf("bottom nav should render a top-bordered bar:\n%s", bar)
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
