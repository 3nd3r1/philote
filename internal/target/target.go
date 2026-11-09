package target

import "github.com/3nd3r1/philote/internal/ir"

// File is one generated file, with a path relative to the part's output dir.
type File struct {
	Path string
	Data []byte
}

// Slot is a hole in a generated file that a model fills, such as a step body.
type Slot struct {
	Node   string // the node it belongs to, like api/record_swipe
	File   string // path of the file that holds it
	Marker string // the line in File that the body replaces
	Prompt string // everything the model is told, built only from IR
}

// Output is what a target generates for one part.
type Output struct {
	Files []File
	Slots []Slot
}

// Target builds one kind of part from IR. Generate must be deterministic:
// the same IR always gives byte-identical files and prompts, in the same order.
type Target interface {
	Name() string
	Kind() ir.PartKind
	// Version changes whenever the target's output for the same IR changes.
	Version() string
	Generate(app *ir.App, part *ir.Part) (Output, error)
	// Verify typechecks a built part in dir and returns problems by file path.
	Verify(dir string) (map[string][]string, error)
}
