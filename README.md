# philote

Philote compiles a strict, typed description of an app into framework code. An LLM is the compiler's backend, and the build system pins its output. Start with the [thesis](ideation/thesis.md).

The project is in the ideation phase. The ideas live as checked items, not as code.

## Layout

```
ideation/
  thesis.md        the pitch and the workflow it enables
  INDEX.md         generated overview; never edit by hand
  principles/      P-xxx  rules the design must follow
  questions/       Q-xxx  open design questions
  decisions/       D-xxx  answers, with the reasoning
  ideas/           I-xxx  things worth exploring
  experiments/     E-xxx  how we find out, with kill criteria
  prior-art/       A-xxx  related work
  archive/         original documents, kept as they were
prototypes/
  frontend-v0/     E-001 step 1: parser, checker, build graph (Go)
tools/
  ideation.py      validator and index generator (Python stdlib only)
```

## Working on the ideation

```sh
make new KIND=question TITLE="Does auth belong in v1?"   # creates the next Q-xxx file
make index                                               # regenerate INDEX.md
make check                                               # validate items, and that INDEX.md is current
```

Each item is a Markdown file with TOML frontmatter between `+++` lines:

```toml
+++
id = "D-001"                # must match the file name
title = "..."               # max 80 chars
status = "proposed"         # allowed values depend on the kind
summary = "..."             # max 200 chars; shown in INDEX.md
created = 2026-11-09        # a TOML date, unquoted
related = ["P-005"]         # any item ids
answers = ["Q-009"]         # decisions only
+++
```

| Kind | Statuses | Extra fields | Required sections |
| --- | --- | --- | --- |
| principle | proposed, accepted, retired | | |
| question | open, answered, dropped | `answered_by` (a decision) | |
| decision | proposed, accepted, rejected, superseded | `answers`, `supersedes` | Decision, Why |
| idea | raw, exploring, adopted, rejected | | |
| experiment | planned, running, done, abandoned | `tests` | Steps, Kill criteria (+ Result when done) |
| prior-art | noted, studied | `url` | |

`make check` enforces:
- Unknown fields are errors. Links must resolve and point to the right kind.
- An answered question names an accepted decision, and that decision lists the question in `answers`. An accepted decision's questions must be marked answered.
- Leftover TODOs are errors.

## Rules

- Anything an AI decides goes in as `proposed`. Only a person moves it to `accepted`. INDEX.md lists proposed items under "Waiting for review".
- Don't delete items whose thinking changed. Reject, retire or supersede them, so the reasoning stays.
- Never edit `INDEX.md` by hand.

## Prototype

```sh
cd prototypes/frontend-v0
go test ./...
go run ./cmd/specc graph examples/blog/blog.spec
```
