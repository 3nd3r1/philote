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
`

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
	if flags.Parse(args[1:]) != nil {
		return 2
	}
	wantArgs := map[string]int{"check": 1, "tree": 1, "expand": 2, "build": 1}
	if n, ok := wantArgs[cmd]; !ok || flags.NArg() != n {
		fmt.Fprint(stderr, usage)
		return 2
	}
	spec := flags.Arg(0)
	app := load(spec, stderr)
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
		return writeBuild(app, *out, stdout, stderr)
	}
	return 0
}

// load runs the front end and prints diagnostics. It returns nil on errors.
func load(spec string, stderr io.Writer) *ir.App {
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
		return nil
	}
	lower.Expand(app)
	return app
}

func writeBuild(app *ir.App, dir string, stdout, stderr io.Writer) int {
	outs, err := build.Generate(app, targets)
	if err == nil {
		err = build.Write(dir, outs)
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
