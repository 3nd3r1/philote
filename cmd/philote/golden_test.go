package main

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/*/want")

// Each testdata/<case>/<case>.philote is compiled, and its build output plus
// the tree and expand views are compared with testdata/<case>/want.
func TestGolden(t *testing.T) {
	dirs, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			spec := filepath.Join(dir, name+".philote")
			got := map[string]string{
				"tree.txt":   runOK(t, "tree", "--depth", "3", spec),
				"expand.txt": runOK(t, "expand", spec, "api"),
			}
			out := t.TempDir()
			runOK(t, "build", "--frozen", "--out", out, spec)
			for path, data := range readTree(t, out) {
				got["build/"+path] = data
			}
			compare(t, filepath.Join(dir, "want"), got)
		})
	}
}

func runOK(t *testing.T, args ...string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("philote %s: exit %d\n%s", strings.Join(args, " "), code, stderr.String())
	}
	return stdout.String()
}

func readTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		rel, _ := filepath.Rel(dir, path)
		files[filepath.ToSlash(rel)] = string(data)
		return err
	})
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return files
}

func compare(t *testing.T, wantDir string, got map[string]string) {
	t.Helper()
	if *update {
		if err := os.RemoveAll(wantDir); err != nil {
			t.Fatal(err)
		}
		for path, data := range got {
			p := filepath.Join(wantDir, path)
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	want := readTree(t, wantDir)
	var paths []string
	for p := range want {
		paths = append(paths, p)
	}
	for p := range got {
		if _, ok := want[p]; !ok {
			paths = append(paths, p)
		}
	}
	slices.Sort(paths)
	for _, p := range paths {
		if want[p] != got[p] {
			t.Errorf("%s differs from the golden file (run go test ./cmd/philote -update and review the diff)\n%s", p, firstDiff(want[p], got[p]))
		}
	}
}

func firstDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := range max(len(w), len(g)) {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return "line " + strconv.Itoa(i+1) + ":\n  want: " + wl + "\n  got:  " + gl
		}
	}
	return ""
}

// A frozen build must fail at the step whose body isn't in the lock, and
// write nothing.
func TestFrozenWithoutBody(t *testing.T) {
	spec := copySpec(t)
	var stdout, stderr bytes.Buffer
	out := filepath.Join(filepath.Dir(spec), "build")
	code := run([]string{"build", "--frozen", "--out", out, spec}, &stdout, &stderr)
	want := "api.philote:13:6: error: step record_swipe has no body in philote.lock for this spec"
	if code != 1 || !strings.Contains(stderr.String(), want) {
		t.Errorf("exit %d, want 1 and %q in:\n%s", code, want, stderr.String())
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Errorf("a failed frozen build wrote %s", out)
	}
}

// copySpec copies the horse example's spec, without its lock, into a temp
// dir and returns the path of its main file.
func copySpec(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"neighbr.philote", "api.philote"} {
		data, err := os.ReadFile(filepath.Join("testdata/neighbr", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "neighbr.philote")
}
