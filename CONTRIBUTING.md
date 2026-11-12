# Contributing

Design notes live in [docs/ideation](docs/ideation/INDEX.md). Start with the [thesis](docs/ideation/thesis.md).

## Develop

```sh
nix-shell
make setup   # once: enables the pre-commit hook
make check   # lint, tests, architecture rules, ideation
make smoke   # build the neighbr test spec, run both apps, send real requests (needs network)
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
