+++
id = "D-007"
title = "Store ideation as typed, validated items with a generated index"
status = "accepted"
summary = "One Markdown file per item with strict TOML frontmatter, typed links between items, a validator, and a generated INDEX.md that must stay current."
created = 2026-11-09
related = ["P-001", "P-008", "I-008"]
+++

## Decision

Ideation lives in `ideation/` as one Markdown file per item. Kinds: principle, question, decision, idea, experiment, prior art. Each file starts with TOML frontmatter that has a fixed schema per kind. `tools/ideation.py check` validates it, and `INDEX.md` is generated from it.

## Why

It uses the project's own principles on its own notes:

- **Strict where it matters, prose where it helps.** Frontmatter is typed and length-capped; unknown keys are errors. The body is free prose.
- **Typed links that are checked.** A decision `answers` questions; an experiment `tests` ideas. Links must resolve, and questions and decisions must agree (an answered question points to an accepted decision that lists it).
- **Generated output is never edited.** `INDEX.md` is rebuilt from the items, and the check fails if it's stale, the same idea as `--frozen`.
- **AI writes, people review.** Anything an AI decides goes in as `proposed`, and INDEX.md lists those under "Waiting for review". Only a person moves it to `accepted`.
- **No dependencies.** TOML frontmatter parses with Python's standard library (tomllib), so the check runs anywhere with Python 3.11+.

## Alternatives considered

- One big Markdown doc: easy to write, impossible to validate, and it drifts.
- Pure YAML or TOML files: validatable, but bad for prose.
- Notion or Claude Docs: good for discussion, not checkable in the repo.
- The project's own DSL: the language isn't designed yet. Kept as I-008.
