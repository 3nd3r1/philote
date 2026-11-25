---
id: TASK-19
title: Template the 8 built-in component kinds in nextjs target
status: To Do
assignee: []
created_date: '2026-11-25T10:05:00Z'
labels: []
milestone: m-6
dependencies:
  - TASK-18
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-6 and decision-7 replace model-written frontend component bodies with 8 deterministic built-in component templates (card, detail, table, list, hero, callout, picker, deck) rendered from fields by type. A component kind is the single dispatch point so a future milestone can add a model-written kind without reworking pages.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 nextjs renders all 8 component kinds (card, detail, table, list, hero, callout, picker, deck) with type-aware field rendering and zero frontend model slots
- [ ] #2 Frontend prompts, component body verifier, and unused model-slot plumbing in internal/target/nextjs are removed cleanly
- [ ] #3 board, library, neighbr, and booking specs and golden outputs are updated to : <kind> and theme: with zero frontend entries in philote.lock
- [ ] #4 make check passes (golangci-lint, archtest, and full go test ./... including Next.js builds)
<!-- AC:END -->
