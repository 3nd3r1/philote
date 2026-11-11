package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

// nameLine is package.json's name, which differs per app but doesn't change
// what gets installed.
var nameLine = regexp.MustCompile(`(?m)^\s*"name": .*\n`)

// linkNodeModules gives dir the dependencies listed in pkgFile, so tests
// run the real tsc. They're installed once per dependency list into the
// user cache, which needs the network the first time.
func linkNodeModules(t *testing.T, pkgFile, dir string) {
	t.Helper()
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("needs npm")
	}
	pkg, err := os.ReadFile(pkgFile)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(nameLine.ReplaceAll(pkg, nil))
	deps := filepath.Join(cache, "philote", "test-deps", hex.EncodeToString(sum[:8]))
	if _, err := os.Stat(filepath.Join(deps, "node_modules")); err != nil {
		installDeps(t, deps, pkg)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(deps, "node_modules"), filepath.Join(dir, "node_modules")); err != nil {
		t.Fatal(err)
	}
}

// installDeps installs into a temp dir and renames it into place, so an
// interrupted install never leaves a half-filled cache.
func installDeps(t *testing.T, dir string, pkg []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dir), "install-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	if err := os.WriteFile(filepath.Join(tmp, "package.json"), pkg, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("npm", "install", "--silent", "--no-audit", "--no-fund")
	cmd.Dir = tmp
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("npm install: %v\n%s", err, out)
	}
	if err := os.Rename(tmp, dir); err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatal(err)
	}
}
