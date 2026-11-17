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

// bodyReturn matches a return at the top level of a component body, which
// the template indents by four spaces; cardRoot, one that returns a Card.
var (
	bodyReturn  = regexp.MustCompile(`(?m)^    return\b`)
	cardRoot    = regexp.MustCompile(`^return\s*\(?\s*<Card[\s>]`)
	contentCall = regexp.MustCompile(`\bcontent\(`)
)

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
	if err := checkLayouts(dir, problems); err != nil {
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

// checkBodies reports each component body whose root isn't a Card. The
// template joins the actions bar to that card, outside the body (D-020).
func checkBodies(dir string, problems map[string][]string) error {
	paths, err := filepath.Glob(filepath.Join(dir, "src/components/common/*.tsx"))
	if err != nil {
		return err
	}
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, body, ok := strings.Cut(string(src), "Body({")
		if !ok {
			continue
		}
		returns := bodyReturn.FindAllStringIndex(body, -1)
		if len(returns) > 0 && cardRoot.MatchString(body[returns[len(returns)-1][0]+4:]) {
			continue
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		file := filepath.ToSlash(rel)
		problems[file] = append(problems[file], "the body must end by returning one <Card> as its root, with nothing around it")
	}
	return nil
}

// checkLayouts reports each page layout that doesn't render the page's
// content exactly once: the template's data and forms live there.
func checkLayouts(dir string, problems map[string][]string) error {
	app := filepath.Join(dir, "src/app")
	return filepath.WalkDir(app, func(path string, e os.DirEntry, err error) error {
		if err != nil || e.IsDir() || e.Name() != "page.tsx" {
			return err
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, body, ok := strings.Cut(string(src), "function PageBody({")
		if !ok || len(contentCall.FindAllString(body, -1)) == 1 {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		file := filepath.ToSlash(rel)
		problems[file] = append(problems[file], "the layout must call content(className) exactly once")
		return nil
	})
}
