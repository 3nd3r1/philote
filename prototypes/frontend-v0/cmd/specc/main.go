package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"specc/internal/graph"
	"specc/internal/spec"
)

const compilerVersion = "0.0.1"

func usage() {
	fmt.Fprint(os.Stderr, `usage: specc <command> [flags] <file.spec>

commands:
  check   parse and type-check the source
  fmt     print the canonical form of the source
  graph   print build nodes, dependencies, interface hashes and cache keys
`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	model := fs.String("model", "claude-opus-5-5", "model used for generated nodes")
	prompt := fs.String("prompt-version", "1", "prompt template version")
	fs.Parse(os.Args[2:])
	if fs.NArg() != 1 {
		usage()
	}
	path := fs.Arg(0)
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	prog, diags := spec.Load(path, string(src))
	if len(diags.List) > 0 {
		fmt.Fprint(os.Stderr, spec.Render(path, string(src), diags.Sorted()))
	}
	if prog == nil {
		os.Exit(1)
	}

	switch cmd {
	case "check":
		f := prog.File
		fmt.Printf("ok: %d entities, %d pages, %d components\n", len(f.Entities), len(f.Pages), len(f.Components))
	case "fmt":
		fmt.Print(spec.FormatFile(prog.File))
	case "graph":
		nodes := graph.Build(prog, graph.Config{CompilerVersion: compilerVersion, PromptVersion: *prompt, Model: *model})
		fmt.Printf("%-22s %-9s %-10s %-10s %s\n", "NODE", "MODE", "IFACE", "KEY", "DEPS")
		for _, n := range nodes {
			fmt.Printf("%-22s %-9s %-10s %-10s %s\n", n.ID, n.Mode, n.IfaceHash[:8], n.Key[:8], strings.Join(n.Deps, " "))
		}
	default:
		usage()
	}
}
