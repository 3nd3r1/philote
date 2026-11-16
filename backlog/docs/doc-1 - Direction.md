---
id: doc-1
title: Direction
type: other
created_date: '2026-11-16T12:10:00Z'
updated_date: '2026-11-16T12:19:00Z'
---
Owned by the user. Agents read it and don't edit it.

## Goal

A language that can build anything, between text and code: more exact than a spec, shorter than code, and deterministic. The same source always builds the same program.

## Where we are (2026-11-16)

Philote builds CRUD apps: a Hono backend and a Next.js frontend, with model-written bodies pinned in `philote.lock`. Two fresh builds differ only in model bodies; with the lock they don't differ at all (E-003). But the language's nouns (entity, resource, page) can only describe CRUD.

## Milestones, in order

1. **Neighbr demo video.** About 60 seconds. First 30: neighbr is created from its spec, fast and easy to follow, and the first seconds hook. Last 30: one thing in the spec changes, the app rebuilds, and it looks exactly the same except that one part, the only thing regenerated. The app looks great, not like the plain dark default.
2. **CLI tool.** No UI, no database. Finds out what the language needs beyond CRUD nouns.
3. **Game server.** State, time, events.
4. **Data pipeline.** External APIs, effects, retries.
5. **Editor UI.** Rich client state that list, detail and form pages can't express.
6. **CRUD as a library.** Today's language rebuilt on whatever the general milestones produce.

## Parked

- Auth (Q-004): when a milestone needs users.
- A second framework (E-001 step 5).
- Native blocks (I-006), an explain command (I-004), checking that a component renders every shown field (I-010), AI writing the source (I-003).
