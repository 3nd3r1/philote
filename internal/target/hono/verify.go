package hono

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// tscError matches tsc's plain output: src/steps/x.ts(12,7): error TS2339: ...
var tscError = regexp.MustCompile(`^(\S+)\((\d+),(\d+)\): error (.*)$`)

// Verify installs dependencies if needed and typechecks the project with tsc.
func (Hono) Verify(dir string) (map[string][]string, error) {
	if _, err := os.Stat(filepath.Join(dir, "node_modules")); errors.Is(err, os.ErrNotExist) {
		if out, err := command(dir, "npm", "install", "--silent", "--no-audit", "--no-fund"); err != nil {
			return nil, fmt.Errorf("npm install in %s: %w\n%s", dir, err, out)
		}
	}
	out, err := command(dir, "npx", "--no-install", "tsc", "--pretty", "false")
	problems := map[string][]string{}
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
