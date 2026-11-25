package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Built-in component kinds render from templates only: zero model calls, no
// component pins in philote.lock, and changing a component's kind rebuilds
// cleanly with --frozen.
func TestTemplatedComponents(t *testing.T) {
	spec := copyLibrary(t)
	out := filepath.Join(filepath.Dir(spec), "build")
	r := useReplies(t, nil)
	runOK(t, "build", "--out", out, spec)
	if len(r.prompts) != 0 {
		t.Fatalf("templated frontend should make 0 model calls; got %v", r.prompts)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(spec), "philote.lock")); !os.IsNotExist(err) {
		t.Errorf("library has no backend steps, so philote.lock should not be written")
	}
	shelf := readFile(t, filepath.Join(out, "web/src/components/common/book-shelf.tsx"))
	if !strings.Contains(shelf, "<table") {
		t.Errorf("BookShelf (: table) should render a <table>:\n%s", shelf)
	}

	src := readFile(t, spec)
	edited := strings.Replace(src, "component BookShelf(books: list of Book): table", "component BookShelf(books: list of Book): list", 1)
	if edited == src {
		t.Fatal("BookShelf header moved; update the test")
	}
	if err := os.WriteFile(spec, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	runOK(t, "build", "--frozen", "--out", out, spec)
	shelf = readFile(t, filepath.Join(out, "web/src/components/common/book-shelf.tsx"))
	if !strings.Contains(shelf, "<ul") {
		t.Errorf("BookShelf (: list) should render a <ul>:\n%s", shelf)
	}
}

func TestRedoErrors(t *testing.T) {
	spec := copyLibrary(t)
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"build", "--redo", "api/nope", spec}, 1, "there is no step at that path"},
		{[]string{"build", "--frozen", "--redo", "api/nope", spec}, 2, "can't be used with --frozen"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(tc.args, &stdout, &stderr); code != tc.code || !strings.Contains(stderr.String(), tc.want) {
			t.Errorf("philote %s: exit %d, want %d and %q in:\n%s", strings.Join(tc.args, " "), code, tc.code, tc.want, stderr.String())
		}
	}
}

// copyLibrary copies the library example's spec into a temp dir.
func copyLibrary(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/library/library.philote")
	if err != nil {
		t.Fatal(err)
	}
	spec := filepath.Join(t.TempDir(), "library.philote")
	if err := os.WriteFile(spec, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return spec
}
