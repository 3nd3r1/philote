package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/3nd3r1/philote/internal/target/nextjs"
)

// Each case's frontend is typechecked with the real tsc, so pages,
// components and reads can't drift from the client and types they use.
func TestFrontendsTypecheck(t *testing.T) {
	dirs, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			pkg := filepath.Join(dir, "want/build/web/package.json")
			if _, err := os.Stat(pkg); err != nil {
				t.Skip("no frontend")
			}
			out := t.TempDir()
			linkNodeModules(t, pkg, filepath.Join(out, "web"))
			runOK(t, "build", "--frozen", "--out", out, filepath.Join(dir, name+".philote"))
			problems, err := nextjs.Nextjs{}.Verify(filepath.Join(out, "web"))
			if err != nil {
				t.Fatal(err)
			}
			for file, list := range problems {
				t.Errorf("%s: %v", file, list)
			}
		})
	}
}
