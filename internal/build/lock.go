package build

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/3nd3r1/philote/internal/diag"
)

// lockHeader opens every philote.lock.
const lockHeader = `# Step bodies written by a model and pinned by philote build (D-015).
# Commit this file and review its diff like code. Don't edit it by hand.
`

// pin is one step body in philote.lock.
type pin struct {
	Node  string // like api/record_swipe
	Key   string
	Model string
	Body  string // dedented, no trailing newline
}

// key pins a body to exactly what produced it: the model and the prompt.
// The prompt is built only from the spec and the target, so it covers
// everything else.
func key(model, prompt string) string {
	sum := sha256.Sum256([]byte(model + "\n" + prompt))
	return hex.EncodeToString(sum[:])
}

// readLock reads the lock at path; a missing lock is empty. Malformed lines
// are reported in d.
func readLock(path string, d *diag.List) (map[string]pin, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]pin{}, nil
	}
	if err != nil {
		return nil, err
	}
	return parseLock(filepath.Base(path), string(data), d), nil
}

func parseLock(file, text string, d *diag.List) map[string]pin {
	pins := map[string]pin{}
	var cur *pin
	var body []string
	done := func() {
		if cur != nil {
			cur.Body = strings.Join(body, "\n")
			pins[cur.Node] = *cur
		}
		body = nil
	}
	for i, line := range strings.Split(text, "\n") {
		word, rest, _ := strings.Cut(line, " ")
		switch {
		case cur != nil && (line == "|" || strings.HasPrefix(line, "| ")):
			body = append(body, strings.TrimPrefix(strings.TrimPrefix(line, "|"), " "))
		case line == "" || strings.HasPrefix(line, "#"):
		case word == "step" && rest != "":
			done()
			cur = &pin{Node: rest}
		case cur != nil && word == "key" && rest != "":
			cur.Key = rest
		case cur != nil && word == "model" && rest != "":
			cur.Model = rest
		default:
			d.Errorf(diag.Pos{File: file, Line: i + 1, Col: 1}, "can't read this line of the lock")
			d.Hint("delete %s and build again to have every step body written fresh", file)
			return nil
		}
	}
	done()
	return pins
}

// printLock writes pins sorted by node, so the same pins give the same bytes.
func printLock(pins map[string]pin) []byte {
	nodes := make([]string, 0, len(pins))
	for n := range pins {
		nodes = append(nodes, n)
	}
	slices.Sort(nodes)
	var b strings.Builder
	b.WriteString(lockHeader)
	// A body line is "| " plus the line, or a bare "|" when empty, so the
	// lock has no trailing spaces.
	for _, n := range nodes {
		p := pins[n]
		b.WriteString("\nstep " + p.Node + "\nkey " + p.Key + "\nmodel " + p.Model + "\n")
		for _, line := range strings.Split(p.Body, "\n") {
			if line == "" {
				b.WriteString("|\n")
			} else {
				b.WriteString("| " + line + "\n")
			}
		}
	}
	return []byte(b.String())
}
