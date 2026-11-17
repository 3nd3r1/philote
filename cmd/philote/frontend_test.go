package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/3nd3r1/philote/internal/target/nextjs"
)

// Each case's frontend is typechecked with the real tsc, so pages,
// components and reads can't drift from the client and types they use. Then
// it gets a production build, which is the only place Next reports what
// blocks prerendering, like a URL hook outside Suspense.
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
			nextBuild(t, pkg, filepath.Join(dir, name+".philote"), name)
		})
	}
}

// nextBuild builds the frontend for production inside the cached install,
// since Turbopack doesn't follow a node_modules symlinked from outside the
// project; Node finds the install by walking up from the app.
func nextBuild(t *testing.T, pkg, spec, name string) {
	t.Helper()
	out := filepath.Join(testDeps(t, pkg), "apps", name)
	if err := os.RemoveAll(out); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(out) })
	runOK(t, "build", "--frozen", "--out", out, spec)
	cmd := exec.Command("npx", "--no-install", "next", "build")
	cmd.Dir = filepath.Join(out, "web")
	cmd.Env = append(os.Environ(), "NEXT_TELEMETRY_DISABLED=1")
	if log, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("next build: %v\n%s", err, log)
	}
}
