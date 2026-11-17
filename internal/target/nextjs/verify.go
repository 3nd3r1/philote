package nextjs

import (
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
	bodyReturn = regexp.MustCompile(`(?m)^    return\b`)
	cardRoot   = regexp.MustCompile(`^return\s*\(?\s*<Card[\s>]`)
)

// Verify checks the theme, installs dependencies if needed and typechecks
// the project. Route types like LayoutProps come from next typegen, so it
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
	out, err := command(dir, "npx", "--no-install", "tsc", "--noEmit", "--pretty", "false")
	for _, line := range strings.Split(out, "\n") {
		if m := tscError.FindStringSubmatch(line); m != nil {
			problems[m[1]] = append(problems[m[1]], fmt.Sprintf("line %s: %s", m[2], m[4]))
		}
	}
	if err != nil && len(problems) == 0 {
		return nil, fmt.Errorf("tsc in %s: %w\n%s", dir, err, out)
	}
	return problems, nil
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
