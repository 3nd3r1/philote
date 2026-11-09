package syntax

import (
	"io/fs"
	"path"

	"github.com/3nd3r1/philote/internal/diag"
)

// Source is a parsed root file with every part body loaded, plus the text
// of each file for rendering diagnostics.
type Source struct {
	Root  *File
	Texts map[string]string
}

// Load parses root and, for each part declared without a body, the file
// named after it next to root (`backend api: hono` loads api.philote).
func Load(fsys fs.FS, root string, d *diag.List) *Source {
	src := &Source{Texts: map[string]string{}}
	src.Root = src.parse(fsys, root, d)
	if src.Root == nil {
		return src
	}
	for _, it := range src.Root.Items {
		app, ok := it.(*App)
		if !ok {
			continue
		}
		for _, it := range app.Items {
			if part, ok := it.(*Part); ok && !part.HasBody {
				src.loadBody(fsys, path.Join(path.Dir(root), part.Name.Name+".philote"), part, d)
			}
		}
	}
	return src
}

func (s *Source) parse(fsys fs.FS, name string, d *diag.List) *File {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		d.Errorf(diag.Pos{File: name}, "can't read %s: %v", name, err)
		return nil
	}
	s.Texts[name] = string(b)
	return Parse(name, string(b), d)
}

func (s *Source) loadBody(fsys fs.FS, name string, part *Part, d *diag.List) {
	if _, err := fs.Stat(fsys, name); err != nil {
		d.Errorf(part.Name.Pos, "%s has no body, and %s doesn't exist", part.Name.Name, name)
		d.Hint("add a { } body here, or create %s", name)
		return
	}
	if f := s.parse(fsys, name, d); f != nil {
		part.Body = f.Items
	}
}
