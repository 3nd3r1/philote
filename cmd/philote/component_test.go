package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Recorded component bodies for the library example. The broken one reads a
// field the props don't have.
const (
	brokenAuthorCard = "```tsx\nreturn <p>{author.nickname}</p>;\n```"
	goodAuthorCard   = "```tsx\nreturn (\n    <Card>\n        <Link href={link.href}>{author.name}</Link>\n        {actions}\n    </Card>\n);\n```"
	goodAuthorView   = "```tsx\nreturn <Card>{author.books.map((book) => book.title).join(\", \")}</Card>;\n```"
	goodBookShelf    = "```tsx\nreturn (\n    <div>\n        {books.map((book) => (\n            <Badge key={book.id}>{book.author.name}</Badge>\n        ))}\n    </div>\n);\n```"
)

// A component body goes through the same loop as a step's: tsc errors go
// back to the model, passing bodies are pinned, a rebuild calls nothing,
// and --redo asks for one body fresh, without the old one.
func TestComponentBodies(t *testing.T) {
	spec := copyLibrary(t)
	out := filepath.Join(filepath.Dir(spec), "build")
	linkNodeModules(t, "testdata/library/want/build/web/package.json", filepath.Join(out, "web"))
	r := useReplies(t, map[string][]string{
		"AuthorCardBody": {brokenAuthorCard, goodAuthorCard},
		"AuthorViewBody": {goodAuthorView},
		"BookShelfBody":  {goodBookShelf},
	})
	runOK(t, "build", "--out", out, spec)
	if n := len(r.prompts["AuthorCardBody"]); n != 2 || !strings.Contains(r.prompts["AuthorCardBody"][1], "nickname") {
		t.Fatalf("AuthorCard was asked %d times, want 2 with the broken body in the retry", n)
	}
	lock := readFile(t, filepath.Join(filepath.Dir(spec), "philote.lock"))
	if !strings.Contains(lock, "component web/AuthorCard\n") || strings.Contains(lock, "nickname") {
		t.Errorf("want the fixed AuthorCard pinned as a component:\n%s", lock)
	}

	r = useReplies(t, map[string][]string{})
	runOK(t, "build", "--out", out, spec)
	if len(r.prompts) != 0 {
		t.Errorf("an unchanged spec called the model: %v", r.prompts)
	}

	r = useReplies(t, map[string][]string{"BookShelfBody": {goodBookShelf}})
	runOK(t, "build", "--redo", "web/BookShelf", "--out", out, spec)
	if len(r.prompts) != 1 || len(r.prompts["BookShelfBody"]) != 1 {
		t.Fatalf("--redo web/BookShelf: want one call for BookShelf, got %v", r.prompts)
	}
	if strings.Contains(r.prompts["BookShelfBody"][0], "An earlier version") {
		t.Errorf("--redo should start fresh, not from the old body")
	}
}

func TestRedoErrors(t *testing.T) {
	spec := copyLibrary(t)
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"build", "--redo", "web/Nope", spec}, 1, "there is no step, component or theme at that path"},
		{[]string{"build", "--frozen", "--redo", "web/BookShelf", spec}, 2, "can't be used with --frozen"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(tc.args, &stdout, &stderr); code != tc.code || !strings.Contains(stderr.String(), tc.want) {
			t.Errorf("philote %s: exit %d, want %d and %q in:\n%s", strings.Join(tc.args, " "), code, tc.code, tc.want, stderr.String())
		}
	}
}

// copyLibrary copies the library example's spec, without its lock, into a
// temp dir.
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
