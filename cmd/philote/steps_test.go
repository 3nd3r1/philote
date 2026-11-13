package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/3nd3r1/philote/internal/model"
)

// Recorded replies for the horse example's steps. The broken one calls a
// function that doesn't exist, so tsc rejects it.
const (
	brokenSwipe = "```ts\nreturn makeSwipe(horse, target, liked);\n```"
	goodSwipe   = "```ts\nif (horse.id === target.id) {\n  fail(\"same_horse\");\n}\n" +
		"return insert(\"Swipe\", { liked, horse_id: horse.id, target_id: target.id }) as Swipe;\n```"
	goodMatch = "Here is the body:\n```ts\n    if (!swipe.liked) {\n      return;\n    }\n" +
		"    insert(\"Match\", { matched_at: new Date().toISOString(), first_id: swipe.horse_id, second_id: swipe.target_id });\n```"
)

// recorded replays replies in order for each step function, and keeps the
// prompts it was sent.
type recorded struct {
	mu      sync.Mutex
	replies map[string][]string // by function name
	prompts map[string][]string
}

func (r *recorded) Name() string { return "recorded" }

func (r *recorded) Write(prompt string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for fn, queue := range r.replies {
		if !strings.Contains(prompt, "function "+fn+"(") {
			continue
		}
		r.prompts[fn] = append(r.prompts[fn], prompt)
		if len(queue) == 0 {
			return "", fmt.Errorf("no recorded reply left for %s", fn)
		}
		r.replies[fn] = queue[1:]
		return queue[0], nil
	}
	return "", errors.New("no recorded reply for this prompt")
}

func useReplies(t *testing.T, replies map[string][]string) *recorded {
	t.Helper()
	r := &recorded{replies: replies, prompts: map[string][]string{}}
	old := newModel
	newModel = func(string) model.Model { return r }
	t.Cleanup(func() { newModel = old })
	return r
}

// A body that fails tsc goes back to the model with the errors; the fixed
// body is pinned. Building again calls nothing, and changing one step asks
// for that step only, starting from its old body.
func TestStepRetryAndReuse(t *testing.T) {
	spec := copySpec(t)
	out := filepath.Join(filepath.Dir(spec), "build")
	linkNodeModules(t, "testdata/neighbr/want/build/api/package.json", filepath.Join(out, "api"))
	r := useReplies(t, map[string][]string{
		"recordSwipe":   {brokenSwipe, goodSwipe},
		"matchIfMutual": {goodMatch},
	})
	runOK(t, "build", "--out", out, spec)

	if n := len(r.prompts["recordSwipe"]); n != 2 {
		t.Fatalf("recordSwipe was asked %d times, want 2", n)
	}
	retry := r.prompts["recordSwipe"][1]
	for _, want := range []string{"did not typecheck", "makeSwipe", "TS2304"} {
		if !strings.Contains(retry, want) {
			t.Errorf("the retry prompt doesn't mention %q:\n%s", want, retry)
		}
	}
	lock := readFile(t, filepath.Join(filepath.Dir(spec), "philote.lock"))
	step := readFile(t, filepath.Join(out, "api/src/steps/record_swipe.ts"))
	for _, text := range []string{lock, step} {
		if strings.Contains(text, "makeSwipe") || !strings.Contains(text, `fail("same_horse");`) {
			t.Errorf("want the fixed body, not the broken one:\n%s", text)
		}
	}
	if !strings.Contains(lock, "| if (!swipe.liked) {\n|   return;\n| }\n") {
		t.Errorf("the match body isn't dedented in the lock:\n%s", lock)
	}

	r = useReplies(t, map[string][]string{})
	runOK(t, "build", "--out", out, spec)
	if len(r.prompts) != 0 {
		t.Errorf("an unchanged spec called the model: %v", r.prompts)
	}

	api := filepath.Join(filepath.Dir(spec), "api.philote")
	edited := strings.Replace(readFile(t, api), "matched now", "matched now, once", 1)
	if err := os.WriteFile(api, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	r = useReplies(t, map[string][]string{"matchIfMutual": {goodMatch}})
	runOK(t, "build", "--out", out, spec)
	if len(r.prompts) != 1 || len(r.prompts["matchIfMutual"]) != 1 {
		t.Fatalf("want one call for matchIfMutual only, got %v", r.prompts)
	}
	if p := r.prompts["matchIfMutual"][0]; !strings.Contains(p, "An earlier version") || !strings.Contains(p, "if (!swipe.liked) {") {
		t.Errorf("the prompt doesn't start from the old body:\n%s", p)
	}
}

// A body that still fails after the retries fails the build at the step's
// line, and isn't pinned.
func TestStepNeverTypechecks(t *testing.T) {
	spec := copySpec(t)
	out := filepath.Join(filepath.Dir(spec), "build")
	linkNodeModules(t, "testdata/neighbr/want/build/api/package.json", filepath.Join(out, "api"))
	r := useReplies(t, map[string][]string{
		"recordSwipe":   {brokenSwipe, brokenSwipe, brokenSwipe},
		"matchIfMutual": {goodMatch},
	})
	var stdout, stderr bytes.Buffer
	code := run([]string{"build", "--out", out, spec}, &stdout, &stderr)
	want := "api.philote:13:6: error: the model's body for step record_swipe still doesn't typecheck after 3 tries"
	if code != 1 || !strings.Contains(stderr.String(), want) || !strings.Contains(stderr.String(), "TS2304") {
		t.Errorf("exit %d, want 1 and %q with the tsc error in:\n%s", code, want, stderr.String())
	}
	if n := len(r.prompts["recordSwipe"]); n != 3 {
		t.Errorf("recordSwipe was asked %d times, want 3", n)
	}
	lock := readFile(t, filepath.Join(filepath.Dir(spec), "philote.lock"))
	if strings.Contains(lock, "record_swipe") || !strings.Contains(lock, "step api/match_if_mutual") {
		t.Errorf("want only the passing step pinned:\n%s", lock)
	}
}

// A step's file only lets it touch the tables it declares, so a body that
// writes another table fails tsc and goes back to the model (D-022).
func TestStepTouchesAnUndeclaredTable(t *testing.T) {
	spec := copySpec(t)
	out := filepath.Join(filepath.Dir(spec), "build")
	linkNodeModules(t, "testdata/neighbr/want/build/api/package.json", filepath.Join(out, "api"))
	stray := "```ts\ninsert(\"Horse\", { name: \"stray\" });\n```"
	r := useReplies(t, map[string][]string{
		"recordSwipe":   {goodSwipe},
		"matchIfMutual": {stray, goodMatch},
	})
	runOK(t, "build", "--out", out, spec)
	prompts := r.prompts["matchIfMutual"]
	if len(prompts) != 2 || !strings.Contains(prompts[1], `"Horse"' is not assignable to parameter of type '"Match"'`) {
		t.Fatalf("want a retry that names the undeclared table, got %d prompts:\n%s", len(prompts), prompts[len(prompts)-1])
	}
	if strings.Contains(prompts[0], "type Horse") {
		t.Errorf("match_if_mutual doesn't reach Horse, but its prompt shows it:\n%s", prompts[0])
	}
}

// Adding a field to Horse rebuilds the step that takes a Horse, and leaves
// the one that doesn't pinned (E-003, D-022).
func TestFieldReachesOnlyItsSteps(t *testing.T) {
	spec := copySpec(t)
	dir := filepath.Dir(spec)
	lock := readFile(t, "testdata/neighbr/philote.lock")
	src := strings.Replace(readFile(t, spec), "  bio: string(max 280)\n", "  bio: string(max 280)\n  color: string(max 20)\n", 1)
	for name, data := range map[string]string{"philote.lock": lock, "neighbr.philote": src} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"build", "--frozen", "--out", filepath.Join(dir, "build"), spec}, &stdout, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "step record_swipe has no body") || strings.Contains(stderr.String(), "match_if_mutual") {
		t.Errorf("exit %d; want only record_swipe to need the model:\n%s", code, stderr.String())
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
