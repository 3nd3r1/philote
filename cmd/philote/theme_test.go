package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A frontend's theme line picks one of the fixed built-in themes with zero
// model calls and no web/theme pin in philote.lock.
func TestFixedTheme(t *testing.T) {
	spec := copyLibrary(t)
	setTheme(t, spec, "warm")
	out := filepath.Join(filepath.Dir(spec), "build")
	r := useReplies(t, nil)
	runOK(t, "build", "--frozen", "--out", out, spec)
	if len(r.prompts) != 0 {
		t.Fatalf("fixed theme should make 0 model calls; got %v", r.prompts)
	}
	if css := readFile(t, filepath.Join(out, "web/src/app/globals.css")); !strings.Contains(css, "--font-heading: var(--font-fraunces);") {
		t.Errorf("globals.css doesn't hold the warm theme:\n%s", css)
	}

	setTheme(t, spec, "editorial")
	runOK(t, "build", "--frozen", "--out", out, spec)
	if css := readFile(t, filepath.Join(out, "web/src/app/globals.css")); !strings.Contains(css, "--font-heading: var(--font-instrument-serif);") {
		t.Errorf("globals.css doesn't hold the editorial theme:\n%s", css)
	}
}

// setTheme puts a theme line in the library spec, or replaces it.
func setTheme(t *testing.T, spec, theme string) {
	t.Helper()
	data := readFile(t, spec)
	if i := strings.Index(data, "    theme: "); i >= 0 {
		end := strings.Index(data[i:], "\n")
		data = data[:i] + data[i+end+1:]
	}
	data = strings.Replace(data, "    uses: api\n", "    uses: api\n    theme: "+theme+"\n", 1)
	if err := os.WriteFile(spec, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
