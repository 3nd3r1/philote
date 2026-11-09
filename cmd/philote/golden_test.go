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
			runOK(t, "build", "--out", out, spec)
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
