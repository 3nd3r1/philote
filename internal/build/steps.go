package build

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/3nd3r1/philote/internal/diag"
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/model"
	"github.com/3nd3r1/philote/internal/target"
)

// retries is how many times a body that doesn't typecheck goes back to the
// model with its errors (D-015).
const retries = 2

// Options says where model-written bodies come from.
type Options struct {
	Lock   string      // path of philote.lock
	Model  model.Model // named in every key, so it's needed even when frozen
	Frozen bool        // never call the model; every body must be in the lock
	Redo   []string    // nodes, like web/HorseCard, to write fresh even if pinned
	Log    io.Writer   // told which nodes go to the model, since that takes a while
}

// node is a spec node whose body a model writes: a step or a component.
type node struct {
	kind, name string
	pos        diag.Pos
}

func (n node) String() string { return n.kind + " " + n.name }

// job is one body the model has to write.
type job struct {
	part     int // index into the outputs
	slot     target.Slot
	key      string
	previous string   // the node's body from the lock, for an older key
	body     string   // the latest try
	err      error    // the model call failed
	problems []string // tsc errors in the latest try
	file     []byte   // the generated file with the latest try in place
}

// Build generates every part, fills its slots from the lock or the model,
// writes it under dir, and typechecks parts that got new bodies. Problems
// the user fixes in the spec or the lock go to d.
func Build(app *ir.App, targets []target.Target, dir string, opts Options, d *diag.List) ([]Output, error) {
	outs, err := generate(app, targets)
	if err != nil {
		return nil, err
	}
	old, err := readLock(opts.Lock, d)
	if err != nil || d.HasErrors() {
		return nil, err
	}
	nodes := bodyNodes(app)
	for _, r := range opts.Redo {
		if _, ok := nodes[r]; !ok {
			return nil, fmt.Errorf("--redo %s: there is no step or component at that path; paths look like web/HorseCard", r)
		}
	}
	pins, jobs := plan(outs, old, opts, nodes, d)
	if d.HasErrors() {
		return nil, nil
	}
	if len(jobs) > 0 {
		fmt.Fprintf(opts.Log, "asking %s for %s\n", opts.Model.Name(), nodeList(jobs))
	}
	failed, err := writeBodies(dir, outs, pins, jobs, opts, nodes)
	if err != nil {
		return nil, err
	}
	for _, j := range failed {
		n := nodes[j.slot.Node]
		d.Errorf(n.pos, "the model's body for %s still doesn't typecheck after %d tries", n, retries+1)
		d.Hint("tsc: %s (see %s)", j.problems[0], filepath.Join(dir, outs[j.part].Part, j.slot.File))
		// Keep the old body: it's the model's starting point next time.
		if p, ok := old[j.slot.Node]; ok {
			pins[j.slot.Node] = p
		}
	}
	if !opts.Frozen {
		err = saveLock(opts.Lock, pins)
	}
	return outs, err
}

// bodyNodes maps each node path, like api/record_swipe, to the node whose
// body a model writes.
func bodyNodes(app *ir.App) map[string]node {
	nodes := map[string]node{}
	for _, p := range app.Parts {
		for _, s := range p.Steps {
			nodes[p.Name+"/"+s.Name] = node{"step", s.Name, s.Pos}
		}
		for _, c := range p.Components {
			nodes[p.Name+"/"+c.Name] = node{"component", c.Name, c.Pos}
		}
	}
	return nodes
}

// plan pins every slot whose key is in the lock and makes a job for the
// rest. A node to redo starts fresh, without its old body.
func plan(outs []Output, old map[string]pin, opts Options, nodes map[string]node, d *diag.List) (map[string]pin, []*job) {
	pins := map[string]pin{}
	var jobs []*job
	for i, o := range outs {
		for _, s := range o.Slots {
			n, ok := nodes[s.Node]
			if !ok {
				panic("build: slot " + s.Node + " is not a step or component in the IR")
			}
			k := key(opts.Model.Name(), s.Prompt)
			p, ok := old[s.Node]
			redo := slices.Contains(opts.Redo, s.Node)
			switch {
			case redo:
				jobs = append(jobs, &job{part: i, slot: s, key: k})
			case ok && p.Key == k:
				pins[s.Node] = p
			case opts.Frozen:
				d.Errorf(n.pos, "%s has no body in philote.lock for this spec", n)
				d.Hint("build without --frozen to have the model write it, then commit philote.lock")
			default:
				jobs = append(jobs, &job{part: i, slot: s, key: k, previous: p.Body})
			}
		}
	}
	return pins, jobs
}

// writeBodies asks the model for each job's body, writes the parts and
// typechecks them, and retries bodies that fail. Bodies that pass are added
// to pins; it returns the jobs that never passed.
func writeBodies(dir string, outs []Output, pins map[string]pin, jobs []*job, opts Options, nodes map[string]node) ([]*job, error) {
	m := opts.Model
	for try := 0; ; try++ {
		ask(m, jobs)
		bodies, err := latest(pins, jobs)
		if err != nil {
			return nil, err
		}
		filled := fill(outs, bodies)
		if err := write(dir, filled); err != nil || len(jobs) == 0 {
			return nil, err
		}
		failed, err := verify(dir, filled, jobs)
		if err != nil {
			return nil, err
		}
		for _, j := range jobs {
			if !slices.Contains(failed, j) {
				pins[j.slot.Node] = pin{Kind: nodes[j.slot.Node].kind, Node: j.slot.Node, Key: j.key, Model: m.Name(), Body: j.body}
			}
		}
		if len(failed) == 0 || try == retries {
			return failed, nil
		}
		fmt.Fprintf(opts.Log, "%s didn't typecheck; asking again with the errors\n", nodeList(failed))
		jobs = failed
	}
}

// nodeList names the jobs' nodes, like api/record_swipe, web/HorseCard.
func nodeList(jobs []*job) string {
	names := make([]string, len(jobs))
	for i, j := range jobs {
		names[i] = j.slot.Node
	}
	slices.Sort(names)
	return strings.Join(names, ", ")
}

// latest is every slot's body: pinned, or the latest try.
func latest(pins map[string]pin, jobs []*job) (map[string]string, error) {
	bodies := map[string]string{}
	for node, p := range pins {
		bodies[node] = p.Body
	}
	for _, j := range jobs {
		if j.err != nil {
			return nil, fmt.Errorf("%s: %w", j.slot.Node, j.err)
		}
		bodies[j.slot.Node] = j.body
	}
	return bodies, nil
}

// ask calls the model for every job at once; each call takes seconds.
func ask(m model.Model, jobs []*job) {
	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Go(func() {
			reply, err := m.Write(j.prompt())
			j.body, j.err = bodyOf(reply), err
		})
	}
	wg.Wait()
}

// prompt is the slot's prompt plus what the model should start from. The
// additions aren't part of the key: the same spec gives the same key
// whatever its history (D-015).
func (j *job) prompt() string {
	var b strings.Builder
	b.WriteString(j.slot.Prompt)
	if j.previous != "" {
		b.WriteString("\nAn earlier version of this function had the body below. Keep what still fits and change only what the description above needs.\n")
		b.WriteString("```ts\n" + j.previous + "\n```\n")
	}
	if j.problems != nil {
		b.WriteString("\nYour last reply did not typecheck. The file with it in place:\n")
		b.WriteString("```ts\n" + string(j.file) + "```\n")
		b.WriteString("tsc reported:\n" + strings.Join(j.problems, "\n") + "\nReply with the fixed statements.\n")
	}
	return b.String()
}

// bodyOf takes the code block out of a reply and dedents it, so the lock
// holds a body the same way however the model indented it.
func bodyOf(reply string) string {
	text := reply
	if _, after, ok := strings.Cut(reply, "```"); ok {
		_, code, _ := strings.Cut(after, "\n") // drops the language tag
		text, _, _ = strings.Cut(code, "```")
	}
	var lines []string
	indent := -1
	for _, l := range strings.Split(text, "\n") {
		l = strings.TrimRight(l, " \t\r")
		if l != "" && (indent < 0 || len(l)-len(strings.TrimLeft(l, " \t")) < indent) {
			indent = len(l) - len(strings.TrimLeft(l, " \t"))
		}
		lines = append(lines, l)
	}
	for i, l := range lines {
		if l != "" {
			lines[i] = l[indent:]
		}
	}
	return strings.Trim(strings.Join(lines, "\n"), "\n")
}

// fill puts each slot's body in place of its marker line, indented like it.
func fill(outs []Output, bodies map[string]string) []Output {
	filled := slices.Clone(outs)
	for i, o := range filled {
		files := slices.Clone(o.Files)
		for _, s := range o.Slots {
			k := slices.IndexFunc(files, func(f target.File) bool { return f.Path == s.File })
			if k < 0 || bytes.Count(files[k].Data, []byte(s.Marker+"\n")) != 1 {
				panic("build: slot " + s.Node + " has no single marker in " + s.File)
			}
			indent := s.Marker[:len(s.Marker)-len(strings.TrimLeft(s.Marker, " \t"))]
			lines := strings.Split(bodies[s.Node], "\n")
			for n, l := range lines {
				if l != "" {
					lines[n] = indent + l
				}
			}
			files[k].Data = bytes.Replace(files[k].Data, []byte(s.Marker+"\n"), []byte(strings.Join(lines, "\n")+"\n"), 1)
		}
		filled[i].Files = files
	}
	return filled
}

// verify typechecks each part with new bodies and returns the jobs whose
// file has errors. Errors in any other file mean the target or the lock is
// broken, and no retry fixes that.
func verify(dir string, filled []Output, jobs []*job) ([]*job, error) {
	var failed []*job
	for i, o := range filled {
		mine := slices.DeleteFunc(slices.Clone(jobs), func(j *job) bool { return j.part != i })
		if len(mine) == 0 {
			continue
		}
		problems, err := o.target.Verify(filepath.Join(dir, o.Part))
		if err != nil {
			return nil, err
		}
		for _, j := range mine {
			j.problems = problems[j.slot.File]
			delete(problems, j.slot.File)
			if j.problems != nil {
				k := slices.IndexFunc(o.Files, func(f target.File) bool { return f.Path == j.slot.File })
				j.file, failed = o.Files[k].Data, append(failed, j)
			}
		}
		if len(problems) > 0 {
			return nil, fmt.Errorf("%s: tsc found errors outside the new bodies: %v", o.Part, problems)
		}
	}
	return failed, nil
}

// saveLock writes the pins, or removes the lock when no body is left.
func saveLock(lock string, pins map[string]pin) error {
	if len(pins) == 0 {
		if err := os.Remove(lock); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return writeIfChanged(lock, printLock(pins))
}
