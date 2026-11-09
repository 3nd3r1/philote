package archtest

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const (
	module       = "github.com/3nd3r1/philote"
	maxFileLines = 400
)

// allowed lists every package and the internal packages it may import.
// Adding a package or an edge here is an architecture change: ask first.
var allowed = map[string][]string{
	"cmd/philote":            {"*"},
	"internal/archtest":      {},
	"internal/diag":          {},
	"internal/ir":            {"internal/diag"},
	"internal/syntax":        {"internal/diag"},
	"internal/check":         {"internal/diag", "internal/syntax", "internal/ir"},
	"internal/lower":         {"internal/diag", "internal/ir"},
	"internal/target":        {"internal/ir"},
	"internal/target/nextjs": {"internal/ir", "internal/target"},
	"internal/target/hono":   {"internal/ir", "internal/target"},
	"internal/model":         {},
	"internal/build":         {"internal/diag", "internal/ir", "internal/lower", "internal/target", "internal/model"},
}

// allowedDeps lists outside modules the user has approved. Keep it empty
// unless they say otherwise.
var allowedDeps = []string{}

func root(t *testing.T) string {
	t.Helper()
	r, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// goFiles maps each package dir (relative to the root) to its .go files.
func goFiles(t *testing.T) map[string][]string {
	t.Helper()
	r := root(t)
	pkgs := map[string][]string{}
	for _, top := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(r, top), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && (d.Name() == "testdata" || d.Name() == "node_modules") {
				return filepath.SkipDir
			}
			if !d.IsDir() && strings.HasSuffix(p, ".go") {
				dir, _ := filepath.Rel(r, filepath.Dir(p))
				pkgs[filepath.ToSlash(dir)] = append(pkgs[filepath.ToSlash(dir)], p)
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	return pkgs
}

func TestEveryPackageIsListed(t *testing.T) {
	for pkg := range goFiles(t) {
		if _, ok := allowed[pkg]; !ok {
			t.Errorf("%s is not in archtest's allowed list; new packages are an architecture change", pkg)
		}
	}
}

func TestEveryPackageHasDoc(t *testing.T) {
	for pkg, files := range goFiles(t) {
		if pkg == "cmd/philote" {
			continue
		}
		if !slices.ContainsFunc(files, func(f string) bool { return filepath.Base(f) == "doc.go" }) {
			t.Errorf("%s has no doc.go saying what the package does", pkg)
		}
	}
}

func TestImports(t *testing.T) {
	fset := token.NewFileSet()
	for pkg, files := range goFiles(t) {
		for _, f := range files {
			parsed, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, imp := range parsed.Imports {
				path, _ := strconv.Unquote(imp.Path.Value)
				checkImport(t, pkg, f, path)
			}
		}
	}
}

func checkImport(t *testing.T, pkg, file, path string) {
	t.Helper()
	rel, internal := strings.CutPrefix(path, module+"/")
	switch {
	case internal:
		ok := allowed[pkg]
		if !slices.Contains(ok, "*") && !slices.Contains(ok, rel) && rel != pkg {
			t.Errorf("%s imports %s, which archtest doesn't allow", file, rel)
		}
	case strings.Contains(strings.Split(path, "/")[0], "."):
		if !slices.ContainsFunc(allowedDeps, func(d string) bool { return strings.HasPrefix(path, d) }) {
			t.Errorf("%s imports outside module %s; new dependencies need the user's OK", file, path)
		}
	}
}

func TestNoUnapprovedRequires(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(root(t), "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.Lines(string(b)) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require") && len(allowedDeps) == 0 {
			t.Errorf("go.mod has %q; new dependencies need the user's OK", line)
		}
	}
}

func TestFileSize(t *testing.T) {
	for _, files := range goFiles(t) {
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			if n := strings.Count(string(b), "\n"); n > maxFileLines {
				t.Errorf("%s has %d lines; the limit is %d, split it by responsibility", f, n, maxFileLines)
			}
		}
	}
}
