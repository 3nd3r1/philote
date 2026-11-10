package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/3nd3r1/philote/internal/build"
	"github.com/3nd3r1/philote/internal/check"
	"github.com/3nd3r1/philote/internal/diag"
	"github.com/3nd3r1/philote/internal/ir"
	"github.com/3nd3r1/philote/internal/lower"
	"github.com/3nd3r1/philote/internal/model"
	"github.com/3nd3r1/philote/internal/syntax"
	"github.com/3nd3r1/philote/internal/target"
	"github.com/3nd3r1/philote/internal/target/hono"
)

const usage = `usage: philote <command> [flags] <app.philote> [path]

commands:
  check                    check the spec
  tree [--depth N]         show the app as a tree
  expand <path>            show a part or resource as the endpoints it becomes, e.g. api/products
  build [--out dir]        generate code (default: build/ next to the spec)
        [--frozen]         never call the model; every step body must be in philote.lock
        [--model name]     the model that writes step bodies (default ` + defaultModel + `)
`

// defaultModel writes step bodies. It's a full name, not an alias like
// "opus", because the name is part of every key in philote.lock.
const defaultModel = "claude-opus-5-5"

var targets = []target.Target{hono.Hono{}}

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
	modelName := flags.String("model", defaultModel, "model for step bodies")
	if flags.Parse(args[1:]) != nil {
		return 2
	}
	wantArgs := map[string]int{"check": 1, "tree": 1, "expand": 2, "build": 1}
	if n, ok := wantArgs[cmd]; !ok || flags.NArg() != n {
		fmt.Fprint(stderr, usage)
		return 2
	}
	spec := flags.Arg(0)
	app, texts := load(spec, stderr)
	if app == nil {
		return 1
	}
	switch cmd {
	case "tree":
		fmt.Fprint(stdout, lower.Tree(app, *depth))
	case "expand":
		text, ok := lower.Show(app, flags.Arg(1))
		if !ok {
			fmt.Fprintf(stderr, "nothing at %s\n", flags.Arg(1))
			return 1
		}
		fmt.Fprint(stdout, text)
	case "build":
		if *out == "" {
			*out = filepath.Join(filepath.Dir(spec), "build")
		}
		opts := build.Options{
			Lock:   filepath.Join(filepath.Dir(spec), "philote.lock"),
			Model:  model.Claude{Model: *modelName},
			Frozen: *frozen,
		}
		return writeBuild(app, texts, *out, opts, stdout, stderr)
	}
	return 0
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
