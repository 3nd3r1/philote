---
id: decision-3
title: Frontend primitives come before the demo
date: '2026-11-17T10:33:00Z'
status: accepted
---
## Context

On 2026-11-17 the user looked at neighbr after decision-1 (theme) and decision-2 (model-written page layouts) and judged it far below what prompting Claude directly gives. The cause is the frontend language itself: three fixed page kinds (list, record, form), a fixed nav, and the model only filling holes in templates. A swipe deck or a landing page can't be expressed at all.

## Decision

The user chose this on 2026-11-17. A new first milestone, m-6 "Frontend primitives", comes before m-5 "Good enough to demo". The bar: neighbr built by philote is not worse, side by side, than the same description prompted to Claude directly, swipe deck and landing page included. Determinism still comes from the lock.

## Consequences

doc-1's milestone order changes (the user asked the agent to make that edit). TASK-3's remaining work and decision-2 are folded into the primitives design. TASK-4 and TASK-5 wait for m-5. The primitives are new language concepts, so their design goes to the user as a decision before any code.

