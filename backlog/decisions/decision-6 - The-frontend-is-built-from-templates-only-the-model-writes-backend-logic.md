---
id: decision-6
title: The frontend is built from templates only; the model writes backend logic
date: '2026-11-24T11:46:00Z'
status: accepted
---
## Context

m-6 asked that philote's frontend be no worse, side by side, than the same app prompted to Claude directly. On neighbr philote lost, and the booking app looked bad. Each model-written component is written alone, without the app's description, and under rules that ban its own words, so it can't match an app Claude writes whole. Giving the model more freedom in the frontend also works against the main bet: the same spec builds the same app.

## Decision

The user chose this on 2026-11-24 ("templated only but keep it modular so a future milestone could come back to tackle this"). The frontend is built from templates only: a component is one of a fixed set of built-in kinds that the spec picks, rendered from its fields by type, and the look is one of a few fixed themes. The model writes only backend step bodies, pinned in philote.lock. Philote's frontend is for someone who wants a clean, predictable UI from a spec, not one as beautiful as prompting Claude. A component's kind is the one place its body comes from, so a later milestone can add a model-written kind without reworking pages.

## Consequences

Supersedes decision-1 (the look line written into a pinned theme) and the model-written parts of decision-4 (component bodies, the nav). decision-5 lapses: the side-by-side tasks TASK-11, TASK-12 and TASK-13 are archived, and the booking and recipe specs stay as test apps. The demo shows determinism: a spec becomes a working app, a rebuild is byte-identical, and a one-line change gives a small diff. Neighbr's swipe deck needs a built-in deck kind or leaves the demo. doc-1's milestone 1 bar and demo wording are the user's to update.

