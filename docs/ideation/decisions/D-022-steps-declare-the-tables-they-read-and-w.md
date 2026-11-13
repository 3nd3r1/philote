+++
id = "D-022"
title = "Steps declare the tables they read and write"
status = "accepted"
summary = "A step lists the tables it reads and writes. Its prompt shows only the types it reaches, and its file can only touch those tables, so tsc refuses a body that strays."
created = 2026-11-13
answers = ["Q-019"]
related = ["D-015", "D-021", "E-003", "P-001", "P-004", "P-005", "P-007"]
+++

## Decision

A step lists the tables it touches:

```
step record_swipe(horse: Horse, target: Horse, liked: bool) -> swipe: Swipe {
  fails: same_horse, already_swiped
  writes: Swipe
  intent: "store the swipe; a horse can't swipe itself, or swipe the same horse twice"
}

step match_if_mutual(swipe: Swipe) {
  reads: Swipe
  writes: Match
  intent: "if this swipe is a like and the target already liked the swiper, create a Match for the two, matched now"
}
```

- `reads` lists tables the body only reads. `writes` lists tables it inserts into, updates or removes from, and may also read. Both are optional. A table can't be in both, or twice in one.
- Parameters and the result don't grant access. A step that takes a Swipe and looks up other swipes declares `reads: Swipe`.
- The prompt shows the types of the parameters, the result, and the declared tables, plus the enums their fields use. Nothing else. Adding a field to Horse no longer changes `match_if_mutual`'s prompt or key.
- The step's file narrows the store to those tables. `all` and `find` take only declared table names, and `insert`, `update` and `remove` only written ones. A body that touches another table fails `tsc` and goes back to the model with the error, like any other type error (D-015). A step that declares nothing gets no store at all.
- The file imports only the types its prompt shows, so a new entity doesn't touch existing step files either.

## Why

E-003 found that adding a field to any entity rewrote every step, because each prompt listed every type. That's the wiggle D-021 is meant to remove.

Declaring is the option where what a step depends on is written in the spec (P-001, P-007) and checked mechanically (P-004): the lock key covers exactly the declared tables, and tsc enforces them. Inferring from parameters can't cover a step that creates a record it has no path to, like `match_if_mutual` creating a Match. The cost is one or two lines per step, and those lines are useful to a reviewer: they say what a step changes without reading its body.

Splitting reads from writes costs nothing extra to check and tells the reviewer which steps change data.

Existing specs need the new lines, and their pinned step bodies get new keys. Rebuilding passes the old body as the starting point, so the bodies should come back close to unchanged.

## Progress

Built on main, 2026-11-13. Rebuilding neighbr gave both step bodies back byte for byte under their new keys. Tests show that a body writing an undeclared table fails tsc and is retried, and that adding a field to Horse now needs the model for `record_swipe` only.
