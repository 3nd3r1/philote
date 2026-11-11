# philote

A compiler: a strict spec goes in, a Next.js frontend and a Hono backend come out. Templates do as much as possible, a model fills only marked slots, and every output is pinned. Read `docs/ideation/thesis.md` and `docs/ideation/INDEX.md` first.

## Layout

- `cmd/philote/`: CLI wiring only. Parse flags, call packages, print.
- `internal/<pkg>/`: one job per package, stated in its `doc.go`.
- `docs/ideation/`: design notes. Format in CONTRIBUTING.md.
- `prototypes/`: throwaway experiments. The code rules below don't apply there.
- `tools/`: repo tooling.

## Architecture

Pipeline: `syntax` → `check` → `ir` → `lower` → `target` → `build`. Only `build` calls `model`.

- Allowed imports between packages are listed in `internal/archtest`. Adding a package or an import edge is an architecture change: ask first.
- `ir` is plain data. Logic lives in the stage that owns it.
- `check` is the only place that decides validity. Later stages assume valid IR and don't re-validate. If they hit something impossible, that's a compiler bug: panic.
- Every error a user sees is a `diag.Diag` at a source position, with a hint when one helps.

## Targets

- A target is a Go type: the node kinds it supports, its version, and IR → files. A node kind it doesn't support is a check error, never a guess.
- Each target has `base/` (a real, runnable project with exact pinned dependency versions) and `templates/` (`text/template` files), both embedded with `embed`.
- Templates hold no logic. Build a plain view struct in Go; the template only prints it.
- The model fills only slots a template marks: step bodies, and frontend component bodies (D-019). Everything else is a template.
- Output is byte-stable: sorted iteration, no timestamps, no absolute paths, formatting done by the template.

## Code

- Standard library only. New dependencies need the user's OK (archtest enforces).
- Files ≤ 400 lines (archtest). Functions ≤ 60 lines, low nesting and complexity (golangci-lint).
- Name things after the domain (node, endpoint, slot), not patterns. No `util`, `common`, `helpers`, `manager`.
- Nothing speculative: no interface with one implementation unless it's a boundary (target, model), no unused options, no dead code, no TODOs.
- Comments say why, not what.
- A little duplication beats the wrong abstraction.

## Tests

- Compiler output is tested with golden files: `testdata/<case>/` holds the input and the expected output tree. `go test ./... -update` rewrites them; review that diff like code.
- Each language feature gets a golden case and at least one error case (message and position).
- Tests run the real pipeline. No mocks, except the model, which is replaced by recorded replies.

## Workflow

- Tools come from `nix-shell`. Run `make setup` once for the git hook. `make check` (lint, tests, archtest, ideation) must pass before every commit.
- Build only what an accepted decision or the current experiment needs.
- Language and architecture changes start as an ideation item (`make new KIND=... TITLE="..."`).
- Decisions and principles you write are `proposed`. Only the user accepts. When they do, set `accepted` and mark each question in `answers` as answered with `answered_by`.
- Don't delete ideation items: reject, retire, supersede or drop them, and say why. Run `make index` after editing. Never edit `INDEX.md` or `docs/ideation/archive/`.
- Keep docs short. Don't add a doc when a rule, test or item would do.
- Commits: `feat:`, `fix:`, `chore:`, `docs:`, `build:`, `test:`, optional scope. Short lowercase subject, no body, no co-authors or trailers. One logical change each.
