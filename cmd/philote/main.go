package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/3nd3r1/philote/internal/build"
	"github.com/3nd3r1/philote/internal/check"
	"github.com/3nd3r1/philote/internal/diag"
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/lower"
	"github.com/3nd3r1/philote/internal/model"
	"github.com/3nd3r1/philote/internal/syntax"
	"github.com/3nd3r1/philote/internal/target"
	"github.com/3nd3r1/philote/internal/target/hono"
	"github.com/3nd3r1/philote/internal/target/nextjs"
)

const usage = `usage: philote <command> [flags] <app.philote> [path]

commands:
  check                    check the spec
  tree [--depth N]         show the app as a tree
  expand <path>            show a part or resource as the endpoints it becomes, e.g. api/products
  build [--out dir]        generate code (default: build/ next to the spec)
        [--frozen]         never call the model; every body must be in philote.lock
        [--model name]     the model that writes step and component bodies (default ` + defaultModel + `)
        [--redo a,b]       have the model write these fresh, e.g. web/HorseCard
`

// defaultModel writes step and component bodies. It's a full name, not an alias like
// "opus", because the name is part of every key in philote.lock.
const defaultModel = "claude-opus-5-5"

var targets = []target.Target{hono.Hono{}, nextjs.Nextjs{}}

// newModel is the model for step bodies. Tests replace it with recorded replies.
var newModel = func(name string) model.Model { return model.Claude{Model: name} }

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	cmd := args[0]
	flags := flag.NewFlagSet(cmd, flag.ContinueOnError)
	flags.SetOutput(stderr)
	depth := flags.Int("depth", 2, "levels to show")
	out := flags.String("out", "", "output directory")
	frozen := flags.Bool("frozen", false, "never call the model")
	modelName := flags.String("model", defaultModel, "model for step and component bodies")
	redo := flags.String("redo", "", "nodes to write fresh, comma-separated")
	positional, err := parseAnywhere(flags, args[1:])
	if err != nil {
		return 2
	}
	wantArgs := map[string]int{"check": 1, "tree": 1, "expand": 2, "build": 1}
	if n, ok := wantArgs[cmd]; !ok || len(positional) != n {
		fmt.Fprint(stderr, usage)
		return 2
	}
	if *redo != "" && *frozen {
		fmt.Fprintln(stderr, "--redo asks the model for new bodies, so it can't be used with --frozen")
		return 2
	}
	spec := positional[0]
	app, texts := load(spec, stderr)
	if app == nil {
		return 1
	}
	switch cmd {
	case "tree":
		fmt.Fprint(stdout, lower.Tree(app, *depth))
	case "expand":
		text, ok := lower.Show(app, positional[1])
		if !ok {
			fmt.Fprintf(stderr, "nothing at %s\n", positional[1])
			return 1
		}
		fmt.Fprint(stdout, text)
	case "build":
		if *out == "" {
			*out = filepath.Join(filepath.Dir(spec), "build")
		}
		return writeBuild(app, texts, *out, buildOptions(spec, *modelName, *frozen, *redo), stdout, stderr)
	}
	return 0
}

// parseAnywhere parses flags before, between and after the positional
// arguments, since `philote build app.philote --frozen` is how people type it.
func parseAnywhere(flags *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for {
		if err := flags.Parse(args); err != nil {
			return nil, err
		}
		if flags.NArg() == 0 {
			return positional, nil
		}
		positional = append(positional, flags.Arg(0))
		args = flags.Args()[1:]
	}
}

func buildOptions(spec, modelName string, frozen bool, redo string) build.Options {
	opts := build.Options{
		Lock:   filepath.Join(filepath.Dir(spec), "philote.lock"),
		Model:  newModel(modelName),
		Frozen: frozen,
	}
	if redo != "" {
		opts.Redo = strings.Split(redo, ",")
	}
	return opts
}

// load runs the front end and prints diagnostics. It returns the app, or
// nil on errors, and the source texts for later diagnostics.
func load(spec string, stderr io.Writer) (*ir.App, map[string]string) {
	d := &diag.List{}
	src := syntax.Load(os.DirFS(filepath.Dir(spec)), filepath.Base(spec), d)
	var app *ir.App
	if !d.HasErrors() {
		kinds := check.Targets{}
		for _, t := range targets {
			kinds[t.Name()] = t.Kind()
		}
		app = check.Run(src, kinds, d)
	}
	fmt.Fprint(stderr, diag.Render(d.Sorted(), src.Texts))
	if d.HasErrors() {
		return nil, nil
	}
	lower.Expand(app)
	return app, src.Texts
}

func writeBuild(app *ir.App, texts map[string]string, dir string, opts build.Options, stdout, stderr io.Writer) int {
	d := &diag.List{}
	outs, err := build.Build(app, targets, dir, opts, d)
	fmt.Fprint(stderr, diag.Render(d.Sorted(), texts))
	if d.HasErrors() {
		return 1
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	for _, o := range outs {
		fmt.Fprintf(stdout, "%s: %d files in %s\n", o.Part, len(o.Files), filepath.Join(dir, o.Part))
	}
	return 0
}
