// Authoring loop for E-002: a model writes a spec from a prose brief and
// fixes checker errors until the spec is clean or it runs out of rounds.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const system = `You write source files in a small, strict spec language.
Use only what the reference allows. Never invent syntax.
If the brief asks for something the language can't express, leave it out and add a line "// gap: <what's missing>" at the top of the file. Don't approximate it with something that means something else.
Reply with the whole file in one fenced block marked spec, and nothing else.`

type reply struct {
	Result     string  `json:"result"`
	IsError    bool    `json:"is_error"`
	CostUSD    float64 `json:"total_cost_usd"`
	DurationMS int     `json:"duration_ms"`
}

// ask is the only place that talks to a model (D-010).
func ask(prompt, dir string) (reply, error) {
	cmd := exec.Command("claude", "-p", "--tools", "", "--system-prompt", system,
		"--output-format", "json", "--no-session-persistence", "--strict-mcp-config")
	cmd.Dir = dir // an empty dir, so no project files or CLAUDE.md leak in
	cmd.Stdin = strings.NewReader(prompt)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return reply{}, fmt.Errorf("claude: %v: %s", err, errb.String())
	}
	var r reply
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		return reply{}, fmt.Errorf("claude output: %v", err)
	}
	if r.IsError {
		return r, fmt.Errorf("claude error: %s", r.Result)
	}
	return r, nil
}

var fence = regexp.MustCompile("(?s)```[a-z]*\\n(.*?)```")

type result struct {
	brief  string
	rounds int
	clean  bool
	cost   float64
	secs   float64
	gaps   []string
	err    error
}

func run(brief, base, specc, outDir, scratch string, maxRounds int) result {
	name := strings.TrimSuffix(filepath.Base(brief), ".md")
	res := result{brief: name}
	text, err := os.ReadFile(brief)
	if err != nil {
		res.err = err
		return res
	}
	dir := filepath.Join(outDir, name)
	os.RemoveAll(dir)
	os.MkdirAll(dir, 0o755)

	first := base + "\n# Brief\n\n" + string(text)
	prompt := first
	for round := 1; round <= maxRounds; round++ {
		res.rounds = round
		start := time.Now()
		r, err := ask(prompt, scratch)
		res.secs += time.Since(start).Seconds()
		if err != nil {
			res.err = err
			return res
		}
		res.cost += r.CostUSD

		spec := r.Result
		if m := fence.FindStringSubmatch(r.Result); m != nil {
			spec = m[1]
		}
		file := fmt.Sprintf("round-%d.spec", round)
		os.WriteFile(filepath.Join(dir, file), []byte(spec), 0o644)

		check := exec.Command(specc, "check", file)
		check.Dir = dir
		out, err := check.CombinedOutput()
		os.WriteFile(filepath.Join(dir, fmt.Sprintf("round-%d.check.txt", round)), out, 0o644)

		res.gaps = nil
		for _, line := range strings.Split(spec, "\n") {
			if g, ok := strings.CutPrefix(strings.TrimSpace(line), "// gap:"); ok {
				res.gaps = append(res.gaps, strings.TrimSpace(g))
			}
		}
		if err == nil {
			res.clean = true
			return res
		}
		prompt = first + "\n# Your previous attempt\n\n```spec\n" + spec +
			"```\n\n# Checker output\n\n" + string(out) +
			"\nFix every error and reply with the whole corrected file."
	}
	return res
}

func main() {
	maxRounds := flag.Int("rounds", 5, "max model calls per brief")
	outDir := flag.String("out", "results", "where to write each round")
	flag.Parse()

	ref, err := os.ReadFile("reference.md")
	must(err)
	example, err := os.ReadFile("example.spec")
	must(err)
	base := string(ref) + "\n# Example\n\n```spec\n" + string(example) + "```\n"

	scratch, err := os.MkdirTemp("", "authoring-loop")
	must(err)
	defer os.RemoveAll(scratch)
	specc := filepath.Join(scratch, "specc")
	build := exec.Command("go", "build", "-o", specc, "./cmd/specc")
	build.Dir = "../frontend-v0"
	build.Stderr = os.Stderr
	must(build.Run())

	briefs, _ := filepath.Glob("briefs/*.md")
	if flag.NArg() > 0 {
		briefs = flag.Args()
	}
	results := make([]result, len(briefs))
	var wg sync.WaitGroup
	for i, b := range briefs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = run(b, base, specc, *outDir, scratch, *maxRounds)
		}()
	}
	wg.Wait()

	var sum strings.Builder
	fmt.Fprintf(&sum, "| Brief | Rounds | Clean | Cost | Time | Gaps reported |\n| --- | --- | --- | --- | --- | --- |\n")
	for _, r := range results {
		gaps := strings.Join(r.gaps, "; ")
		if r.err != nil {
			gaps = "ERROR: " + r.err.Error()
		}
		fmt.Fprintf(&sum, "| %s | %d | %v | $%.3f | %.0fs | %s |\n", r.brief, r.rounds, r.clean, r.cost, r.secs, gaps)
	}
	fmt.Print(sum.String())
	must(os.WriteFile(filepath.Join(*outDir, "summary.md"), []byte(sum.String()), 0o644))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
