package target

import "github.com/3nd3r1/philote/internal/ir"

// File is one generated file, with a path relative to the part's output dir.
type File struct {
	Path string
	Data []byte
}

// Target builds one kind of part from IR. Generate must be deterministic:
// the same IR always gives byte-identical files in the same order.
type Target interface {
	Name() string
	Kind() ir.PartKind
	// Version changes whenever the target's output for the same IR changes.
	Version() string
	Generate(app *ir.App, part *ir.Part) ([]File, error)
}
