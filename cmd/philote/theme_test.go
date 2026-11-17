package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A frontend's theme goes through the same loop as a body: problems go back
// to the model, a passing theme is pinned, and changing the look rewrites
// the theme and the page layouts, never the components.
func TestThemeFromLook(t *testing.T) {
	spec := copyLibrary(t)
	setLook(t, spec, "calm and bookish")
	out := filepath.Join(filepath.Dir(spec), "build")
	linkNodeModules(t, "testdata/library/want/build/web/package.json", filepath.Join(out, "web"))
	good := pinnedTheme(t)
	unreadable := strings.Replace(good, "--muted-foreground: oklch(0.5 0.03 40);", "--muted-foreground: oklch(0.95 0.03 40);", 1)
	if unreadable == good {
		t.Fatal("the pinned neighbr theme changed; update the muted-foreground this test replaces")
	}
	r := useReplies(t, map[string][]string{
		"theme":          {"```css\n" + unreadable + "\n```", "```css\n" + good + "\n```"},
		"AuthorCardBody": {goodAuthorCard},
		"AuthorViewBody": {goodAuthorView},
		"BookCardBody":   {goodBookCard},
		"PageBody":       libraryLayouts(),
	})
	runOK(t, "build", "--out", out, spec)
	if n := len(r.prompts["theme"]); n != 2 || !strings.Contains(r.prompts["theme"][1], "--muted-foreground on --background has contrast") {
		t.Fatalf("the theme was asked %d times, want 2 with the contrast problem in the retry", n)
	}
	lock := readFile(t, filepath.Join(filepath.Dir(spec), "philote.lock"))
	if !strings.Contains(lock, "theme web/theme\n") || strings.Contains(lock, "0.95 0.03 40") {
		t.Errorf("want the readable theme pinned as web/theme:\n%s", lock)
	}
	if css := readFile(t, filepath.Join(out, "web/src/app/globals.css")); !strings.Contains(css, "--font-heading: var(--font-fraunces);") {
		t.Errorf("globals.css doesn't hold the pinned theme:\n%s", css)
	}

	setLook(t, spec, "calm and bookish, but darker")
	r = useReplies(t, map[string][]string{"theme": {"```css\n" + good + "\n```"}, "PageBody": libraryLayouts()})
	runOK(t, "build", "--out", out, spec)
	if len(r.prompts) != 2 || len(r.prompts["theme"]) != 1 || !strings.Contains(r.prompts["theme"][0], "An earlier version") {
		t.Errorf("a new look should ask for the theme, starting from the old one, and the page layouts only; got %v", r.prompts)
	}
}

// setLook puts the frontend's look line in the library spec, or replaces it.
func setLook(t *testing.T, spec, look string) {
	t.Helper()
	data := readFile(t, spec)
	if i := strings.Index(data, "    look: \""); i >= 0 {
		end := strings.Index(data[i:], "\n")
		data = data[:i] + data[i+end+1:]
	}
	data = strings.Replace(data, "    uses: api\n", "    uses: api\n    look: \""+look+"\"\n", 1)
	if err := os.WriteFile(spec, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

// pinnedTheme is the neighbr example's theme, from its lock.
func pinnedTheme(t *testing.T) string {
	t.Helper()
	lock := readFile(t, "testdata/neighbr/philote.lock")
	_, rest, ok := strings.Cut(lock, "theme web/theme\n")
	if !ok {
		t.Fatal("testdata/neighbr/philote.lock has no theme")
	}
	var lines []string
	for _, l := range strings.Split(rest, "\n") {
		switch {
		case l == "|":
			lines = append(lines, "")
		case strings.HasPrefix(l, "| "):
			lines = append(lines, l[2:])
		case len(lines) > 0:
			return strings.Join(lines, "\n")
		}
	}
	return strings.Join(lines, "\n")
}
