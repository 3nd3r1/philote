# philote

A compiler: a strict spec goes in, a Next.js frontend and a Hono backend come out. Templates do as much as possible, a model fills only marked slots, and every output is pinned. Read `docs/thesis.md`, `docs/language.md` and the direction (`backlog doc view doc-1`) first.

## Layout

- `cmd/philote/`: CLI wiring only. Parse flags, call packages, print.
- `internal/<pkg>/`: one job per package, stated in its `doc.go`.
- `docs/`: the thesis and the language reference. `docs/archive/` is frozen design history: read it for why, never edit it.
- `backlog/`: the board, managed with the `backlog` CLI.
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

- Tools come from `nix-shell`, which also puts the pinned `backlog` CLI on the path. Run `make setup` once for the git hook. `make check` (lint, tests, archtest) must pass before every commit.
- The user steers, like a CEO. They own the direction (doc-1): the goal, the milestone order, new language concepts, pivots. Bring them only direction calls, as one question with a recommendation, and record each answer with `backlog decision create`. Only the user edits doc-1.
- Agents own the tasks: create, split, update and close them without asking. Never ask the user to review a task.
- Keep the board small: tasks only for the current milestone, at most five open. No task for exploration or mechanical edits.
- Build only what the current milestone needs.
- A language change updates `docs/language.md` and gets a golden case.
- Keep docs short. Don't add a doc when a rule, test or task would do.
- Commits: `feat:`, `fix:`, `chore:`, `docs:`, `build:`, `test:`, optional scope. Short lowercase subject, no body, no co-authors or trailers. One logical change each.

<!-- BACKLOG.MD GUIDELINES START -->
<!-- backlog.md-instructions-version: 1.55.3 -->
<CRITICAL_INSTRUCTION>

## Backlog.md Workflow

This project uses Backlog.md for task and project management.

**At the beginning of each conversation in this project, run `backlog instructions overview` before answering or taking action. Re-read it only if you have not read it yet in the current conversation.**

Use the overview to decide whether to search, read, create, or update Backlog tasks.

Before task lifecycle actions, read the matching detailed guide:
- `backlog instructions task-creation` before creating or splitting tasks
- `backlog instructions task-execution` before planning, changing status or assignee, adding a plan or implementation notes, or implementing task work
- `backlog instructions task-finalization` before checking acceptance criteria, writing final summaries, or moving tasks to terminal statuses

Use `backlog <command> --help` before running unfamiliar commands. Help shows options, fields, and examples.

Do not edit Backlog task, draft, document, decision, or milestone markdown files directly. Use the `backlog` CLI so metadata, relationships, and history stay consistent.

</CRITICAL_INSTRUCTION>
<!-- BACKLOG.MD GUIDELINES END -->
