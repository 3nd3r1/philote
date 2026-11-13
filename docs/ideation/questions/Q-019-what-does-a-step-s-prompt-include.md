+++
id = "Q-019"
title = "What does a step's prompt include?"
status = "answered"
answered_by = "D-022"
summary = "Every step's prompt lists every entity's type, so adding a field anywhere rewrites every step (E-003). All types, only the ones the step reaches, or the ones it declares?"
created = 2026-11-12
related = ["E-003", "D-015", "D-021", "P-005"]
+++

A step can read and write any table, so its prompt lists every entity's type. That makes the prompt, and the lock key, depend on the whole data model. Adding `color` to Horse rewrites `match_if_mutual`, which takes a Swipe and creates a Match.

## Options

- **All types (now).** Simple, and the model always knows every table. Any data change rewrites every step, which works against D-021.
- **Only what the step reaches.** Its parameters' and result's entities, plus the entities their relations point at. No new syntax. A step that writes a table it can't reach, like `match_if_mutual` creating a Match, would need a relation path to it or wouldn't know its shape.
- **Declared.** A step lists what it touches: `step match_if_mutual(swipe: Swipe) { writes: Match ... }`. Exact and checkable (P-007): the checker can refuse a body that touches an undeclared table. One more line per step.

Declared fits P-001 and D-021 best: what a step depends on is written in the spec, and the lock key covers exactly that.
