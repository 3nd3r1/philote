package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Model writes text for a prompt. It's a boundary: the claude CLI now, the
// API later (D-010), and recorded replies in tests.
type Model interface {
	// Name is the model's full name. It is part of every lock key, so an
	// alias like "sonnet" that moves to a new model would be wrong here.
	Name() string
	Write(prompt string) (string, error)
}

// system replaces Claude Code's own system prompt, so the reply depends on
// our prompt and the model, not on the CLI version or the machine.
const system = "You write code for a compiler. Follow the instructions exactly and reply only with what they ask for."

// timeout bounds one call; a step body takes seconds, so this only catches a hung CLI.
const timeout = 5 * time.Minute

// Claude calls the local claude CLI, which runs on the user's subscription.
type Claude struct {
	Model string
}

func (c Claude) Name() string { return c.Model }

func (c Claude) Write(prompt string) (string, error) {
	// An empty working dir keeps the CLI from reading a project's CLAUDE.md.
	dir, err := os.MkdirTemp("", "philote-model-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "claude", "-p",
		"--model", c.Model,
		"--system-prompt", system,
		"--tools", "",
		"--output-format", "json",
		"--no-session-persistence",
		"--strict-mcp-config",
		"--setting-sources", "",
		"--disable-slash-commands",
	)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(prompt)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if errors.Is(err, exec.ErrNotFound) {
		return "", errors.New("the claude CLI is not installed; install it, or build with --frozen from a complete philote.lock")
	}
	if err != nil {
		return "", fmt.Errorf("claude: %w\n%s", err, stderr.String())
	}
	var reply struct {
		Result  string `json:"result"`
		IsError bool   `json:"is_error"`
	}
	if err := json.Unmarshal(out, &reply); err != nil {
		return "", fmt.Errorf("claude: unreadable reply: %w", err)
	}
	if reply.IsError {
		return "", fmt.Errorf("claude: %s", reply.Result)
	}
	return reply.Result, nil
}
