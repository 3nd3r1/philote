+++
id = "D-015"
title = "Steps: model-written bodies pinned in a readable lockfile"
status = "accepted"
summary = "The model writes only step bodies. Each body is keyed by a hash of its prompt and stored in philote.lock, which is committed. Same spec, zero model calls. Bodies are typechecked and retried."
created = 2026-11-09
answers = ["Q-006"]
related = ["D-014", "D-010", "I-007", "P-003", "P-004"]
+++

## Decision

1. **The target owns the prompt.** For each step, the target writes the file around it (imports, the exact function signature, how errors are raised) and a prompt describing the runtime API, the types involved, the step's signature, its errors and its intent. The model writes only the function body.
2. **The key is a hash of the model id and the prompt.** The prompt is built only from the canonical spec and the target, so it covers everything that can change the output. The model id includes the claude CLI version (D-010).
3. **`philote.lock` holds the bodies themselves, next to the spec, and is committed.** Plain text, one entry per step: node path, key, model, body. A changed step shows up in review as a readable code diff. No separate store.
4. **Builds:**
   - Normal: steps whose key is in the lock are copied, with no model call. Others call the model and are added. Entries for removed steps are dropped.
   - `--frozen` (CI): never call the model; fail if any step's key is missing.
5. **Regeneration starts from the old body** (Q-006). When a step's key changes, the previous body is passed as a reference so the model changes as little as possible. The reference is not part of the key, so the same spec still gives the same key whatever its history.
6. **Bodies are verified.** After writing, the build typechecks the generated project with `tsc`. If a new body fails, its errors go back to the model, up to 2 retries. After that the build fails, pointing at the step's line in the spec.

## Syntax

Following D-014, a step names its result so later steps can use it:

```
step create_order(email: email, lines: [CheckoutLine]) -> order: Order {
  intent: "create the order and one line per item, at the current product price"
}
```

## Scope

Steps can read and write the app's data, and raise their declared errors. Calling outside services (payments, email) needs a native step (I-006) and is out of scope.

## Why

This is the core claim of the project: model output can be pinned and replayed like a Nix build. Keeping the bodies in the lockfile makes the model's work reviewable in the same diff as the spec change that caused it.
