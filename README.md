<img src="docs/assets/logo.png" alt="philote" width="380">

Can we bring Nix-style determinism to AI-driven coding?

When I build web apps and backends with AI, the idea it generates usually makes sense, but each run picks different technologies I have no context on, and the results vary. Philote tries to fix that. You write a strict spec of what the app does. The compiler turns it into a Next.js frontend and a Hono backend, mostly from templates, so the stack is always the same. A model fills only the gaps the templates mark, and its output is pinned in a lockfile the way Nix pins builds, so the same spec always gives the same code.

Status: design. Start with the [thesis](docs/ideation/thesis.md) and the [index](docs/ideation/INDEX.md).

## Develop

```sh
nix-shell
make setup   # once: enables the pre-commit hook
make check   # lint, tests, architecture rules, ideation
```

## Ideation items

One Markdown file per item, with TOML frontmatter. `make new KIND=question TITLE="..."` creates one; `make check` validates them.

```toml
+++
id = "D-001"            # matches the file name
title = "..."           # max 80 chars
status = "proposed"
summary = "..."         # max 200 chars
created = 2026-11-09
related = ["P-005"]
answers = ["Q-009"]     # decisions only
+++
```

| Kind | Statuses | Extra fields | Required sections |
| --- | --- | --- | --- |
| principle | proposed, accepted, retired | | |
| question | open, answered, dropped | `answered_by` | |
| decision | proposed, accepted, rejected, superseded | `answers`, `supersedes` | Decision, Why |
| idea | raw, exploring, adopted, rejected | | |
| experiment | planned, running, done, abandoned | `tests` | Steps, Kill criteria, Result when done |
| prior-art | noted, studied | `url` | |
