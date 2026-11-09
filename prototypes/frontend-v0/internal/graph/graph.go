// Package graph turns a checked program into build nodes with cache keys.
//
// A node's key covers its own canonical source, the interface hashes of the
// nodes it depends on (never their generated code), and the build config.
// Same key => cache hit => no model call.
package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"specc/internal/spec"
)

const KeyVersion = "specc-node-key/1"

type Config struct {
	CompilerVersion string
	PromptVersion   string
	Model           string
}

// Mode says how a node is built.
type Mode string

const (
	Template Mode = "template" // deterministic codegen, no model
	Model    Mode = "model"    // LLM backend
	Native   Mode = "native"   // handwritten, only checked
)

type Node struct {
	ID        string
	Mode      Mode
	Source    string // canonical source of the node
	Iface     string // canonical interface
	IfaceHash string
	Deps      []string
	Key       string
}

func Hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func Build(p *spec.Program, cfg Config) []*Node {
	var nodes []*Node
	byID := map[string]*Node{}
	add := func(id string, mode Mode, src, iface string) {
		n := &Node{ID: id, Mode: mode, Source: src, Iface: iface, IfaceHash: Hash(iface)}
		for d := range p.Deps[id] {
			n.Deps = append(n.Deps, d)
		}
		sort.Strings(n.Deps)
		nodes = append(nodes, n)
		byID[id] = n
	}
	for _, e := range p.File.Entities {
		add(spec.EntityID(e.Name.Name), Template, spec.FormatEntity(e), spec.IfaceEntity(e))
	}
	for _, pg := range p.File.Pages {
		add(spec.PageID(pg.Route.Raw), Model, spec.FormatPage(pg), spec.IfacePage(pg))
	}
	for _, c := range p.File.Components {
		mode := Model
		if c.Native {
			mode = Native
		}
		add(spec.ComponentID(c.Name.Name), mode, spec.FormatComponent(c), spec.IfaceComponent(c))
	}

	app := p.File.App
	for _, n := range nodes {
		var b strings.Builder
		fmt.Fprintf(&b, "%s\ncompiler %s\ntarget %s\n", KeyVersion, cfg.CompilerVersion, app.Target.Name)
		if n.Mode == Model {
			fmt.Fprintf(&b, "model %s\nprompt %s\nstyle %s\n", cfg.Model, cfg.PromptVersion, app.Style.Value)
		}
		fmt.Fprintf(&b, "node %s %s\nsource\n%s", n.ID, n.Mode, n.Source)
		for _, d := range n.Deps {
			fmt.Fprintf(&b, "dep %s %s\n", d, byID[d].IfaceHash)
		}
		n.Key = Hash(b.String())
	}
	return nodes
}
