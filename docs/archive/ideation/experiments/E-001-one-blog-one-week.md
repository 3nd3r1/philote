+++
id = "E-001"
title = "One week, one blog, five checks"
status = "running"
summary = "Find out whether local regeneration and retargeting hold up. Steps 1 to 3 are done in the real compiler. Step 4 holds for model calls; step 5 and the Claude Code comparison are open."
created = 2026-11-09
tests = ["P-005", "P-006", "D-001", "D-003", "D-004"]
related = ["Q-013", "Q-014"]
+++

## Steps

1. Write a blog in the sketch syntax (2 entities, 2 pages, 2 components) and parse it with a quick parser. **Done**, see Progress.
2. Compile to Next.js with one model call per node. Verify with `tsc`, `next build` and a route check. Blocked on Q-013 and Q-014.
3. Add the lockfile and cache. A rebuild with no changes should make zero model calls and produce byte-identical output.
4. Change only `PostCard`. Only `PostCard` should regenerate, and the diff everywhere else should be empty.
5. Retarget the same source to a second framework (SvelteKit or Astro).

## Measures

Model calls per build, wall time, verification failures, lines changed per source edit. Also make one feature change both ways (here, and with Claude Code on a plain repo) and compare.

## Kill criteria

- Step 4 can't stay local, because interface churn keeps breaking neighbouring nodes.
- Step 5 needs target-specific edits to the source.

If either happens, rethink the node boundaries before writing more compiler.

## Progress

**2026-11-09, step 1.** Prototype in `prototypes/frontend-v0` (Go, removed after cfe6ded): lexer, parser, checker, canonical printer, and a build graph with interface hashes and cache keys. It checks the blog clean, has 20 error-case tests with source positions and hints, and `fmt` round-trips the blog byte-for-byte.

Step 4 already holds at the cache-key level. Each edit changes exactly these keys:

| Edit | Rebuilds |
| --- | --- |
| Change PostCard's `shows` | PostCard only |
| Rename PostCard's parameter | PostCard and /blog |
| Add a field to Tag | Tag, Post, PostView |
| Whitespace or comments | nothing |

That shows the keys are local, not that the generated code stays local. Step 2 is needed for that. Things step 1 turned up: Q-011, Q-012, D-001 to D-006.

**2026-11-12, steps 2 to 4, in the real compiler.** The blog became the neighbr and library test specs (D-011), built by `cmd/philote` rather than the prototype. Steps 2 and 3 are done: the Hono backend and Next.js frontend build, `tsc` and a smoke test pass, and bodies are pinned in `philote.lock` (D-015, D-019). Tests show a rebuild with no changes makes zero model calls, and editing one step asks the model for that step only. Step 4 holds for model calls. Nobody has measured lines changed in the output per spec edit yet. Step 5 and the comparison with Claude Code on a plain repo haven't started. D-021 makes that comparison the measure that matters most.
