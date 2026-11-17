package nextjs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// tscError matches tsc's plain output: src/lib/api.ts(12,7): error TS2339: ...
var tscError = regexp.MustCompile(`^(\S+)\((\d+),(\d+)\): error (.*)$`)

// saysItem matches one item of a component's says array.
var saysItem = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)

// Verify checks the theme and the bodies' shape, installs dependencies if
// needed, then typechecks and lints the project. Route types like LayoutProps come from next typegen, so it
// runs first.
func (Nextjs) Verify(dir string) (map[string][]string, error) {
	css, err := os.ReadFile(filepath.Join(dir, "src/app/globals.css"))
	if err != nil {
		return nil, err
	}
	problems := map[string][]string{}
	if p := checkTheme(string(css)); len(p) > 0 {
		problems["src/app/globals.css"] = p
	}
	if err := checkBodies(dir, problems); err != nil {
		return nil, err
	}
	if _, err := os.Stat(filepath.Join(dir, "node_modules")); errors.Is(err, os.ErrNotExist) {
		if out, err := command(dir, "npm", "install", "--silent", "--no-audit", "--no-fund"); err != nil {
			return nil, fmt.Errorf("npm install in %s: %w\n%s", dir, err, out)
		}
	}
	if out, err := command(dir, "npx", "--no-install", "next", "typegen"); err != nil {
		return nil, fmt.Errorf("next typegen in %s: %w\n%s", dir, err, out)
	}
	if err := typecheck(dir, problems); err != nil {
		return nil, err
	}
	if err := lint(dir, problems); err != nil {
		return nil, err
	}
	return problems, nil
}

func typecheck(dir string, problems map[string][]string) error {
	out, err := command(dir, "npx", "--no-install", "tsc", "--noEmit", "--pretty", "false")
	found := false
	for _, line := range strings.Split(out, "\n") {
		if m := tscError.FindStringSubmatch(line); m != nil {
			problems[m[1]] = append(problems[m[1]], fmt.Sprintf("line %s: %s", m[2], m[4]))
			found = true
		}
	}
	if err != nil && !found {
		return fmt.Errorf("tsc in %s: %w\n%s", dir, err, out)
	}
	return nil
}

// eslintResult is the part of ESLint's JSON report that lint reads.
type eslintResult struct {
	FilePath string
	Messages []struct {
		Line     int
		Message  string
		Severity int
	}
}

// lint runs the app's own ESLint config, the one `npm run lint` checks, so
// a body that typechecks but breaks a lint rule goes back to the model too.
func lint(dir string, problems map[string][]string) error {
	cmd := exec.Command("npx", "--no-install", "eslint", "--format", "json", "src")
	cmd.Dir = dir
	out, runErr := cmd.Output()
	var results []eslintResult
	if err := json.Unmarshal(out, &results); err != nil {
		return fmt.Errorf("eslint in %s: %w\n%s", dir, errors.Join(err, runErr), out)
	}
	for _, r := range results {
		rel, err := filepath.Rel(dir, r.FilePath)
		if err != nil {
			return err
		}
		file := filepath.ToSlash(rel)
		for _, m := range r.Messages {
			if m.Severity == 2 {
				problems[file] = append(problems[file], fmt.Sprintf("line %d: %s", m.Line, m.Message))
			}
		}
	}
	return nil
}

func command(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// checkBodies reports each component body, and the nav's, that leaves out
// what the template hands it: every line of says, the link and the actions,
// or the nav's items, active item, home link and toggle.
func checkBodies(dir string, problems map[string][]string) error {
	paths, err := filepath.Glob(filepath.Join(dir, "src/components/common/*.tsx"))
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, navFile)); err == nil {
		paths = append(paths, filepath.Join(dir, navFile))
	}
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		head, body, ok := strings.Cut(string(src), "Body({")
		if !ok {
			continue
		}
		// The body starts after its signature, which names every prop.
		_, body, _ = strings.Cut(body, "React.ReactNode {")
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		file := filepath.ToSlash(rel)
		problems[file] = append(problems[file], bodyProblems(head, body)...)
		if len(problems[file]) == 0 {
			delete(problems, file)
		}
	}
	return nil
}

// bodyProblems checks one body against the wrapper above it.
func bodyProblems(head, body string) []string {
	var out []string
	if _, says, ok := strings.Cut(head, "const says = ["); ok {
		says, _, _ = strings.Cut(says, "] as const;")
		for i := range saysItem.FindAllString(says, -1) {
			if !strings.Contains(body, fmt.Sprintf("says[%d]", i)) {
				out = append(out, fmt.Sprintf("the body must render says[%d]", i))
			}
		}
	}
	if strings.Contains(head, " link={{") && !strings.Contains(body, "link.href") {
		out = append(out, "the body must link to link.href")
	}
	if strings.Contains(head, " actions={") && strings.Count(body, "{actions}") != 1 {
		out = append(out, "the body must render {actions} exactly once")
	}
	if strings.Contains(head, "<NavBody ") {
		out = append(out, navProblems(body)...)
	}
	return out
}

func navProblems(body string) []string {
	var out []string
	for _, want := range []string{"items.map", "active", "home.href", "home.label"} {
		if !strings.Contains(body, want) {
			out = append(out, "the nav must use "+want)
		}
	}
	if strings.Count(body, "{toggle}") != 1 {
		out = append(out, "the nav must render {toggle} exactly once")
	}
	return out
}
