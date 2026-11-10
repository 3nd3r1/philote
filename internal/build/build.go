package build

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/target"
)

// manifest lists the files a build wrote into a part's dir, so the next
// build can remove the ones it no longer generates and leave everything
// else (node_modules, data files) alone.
const manifest = ".philote-files"

// Output is the generated files for one part, and the slots in them that
// a model fills.
type Output struct {
	Part   string
	Files  []target.File
	Slots  []target.Slot
	target target.Target
}

// generate runs each part's target. Targets are looked up by name; check
// has already made sure every part's target exists.
func generate(app *ir.App, targets []target.Target) ([]Output, error) {
	byName := map[string]target.Target{}
	for _, t := range targets {
		byName[t.Name()] = t
	}
	var out []Output
	for _, p := range app.Parts {
		t := byName[p.Target]
		gen, err := t.Generate(app, p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.Name, err)
		}
		out = append(out, Output{Part: p.Name, Files: gen.Files, Slots: gen.Slots, target: t})
	}
	return out, nil
}

// write puts each part's files under dir/<part>/. Unchanged files are not
// rewritten, so their modification times stay put.
func write(dir string, outs []Output) error {
	for _, o := range outs {
		if err := writePart(filepath.Join(dir, o.Part), o.Files); err != nil {
			return err
		}
	}
	return nil
}

func writePart(dir string, files []target.File) error {
	keep := map[string]bool{}
	var list strings.Builder
	for _, f := range files {
		keep[f.Path] = true
		list.WriteString(f.Path + "\n")
		if err := writeIfChanged(filepath.Join(dir, f.Path), f.Data); err != nil {
			return err
		}
	}
	if err := removeStale(dir, keep); err != nil {
		return err
	}
	return writeIfChanged(filepath.Join(dir, manifest), []byte(list.String()))
}

func writeIfChanged(path string, data []byte) error {
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, data) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func removeStale(dir string, keep map[string]bool) error {
	old, err := os.ReadFile(filepath.Join(dir, manifest))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, p := range strings.Split(strings.TrimSpace(string(old)), "\n") {
		// Only ever delete inside dir, even if the manifest was edited.
		if p == "" || keep[p] || !filepath.IsLocal(p) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, p)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}
