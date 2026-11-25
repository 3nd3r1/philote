---
id: TASK-19
title: Template the 8 built-in component kinds in nextjs target
status: Done
assignee:
  - '@agy'
created_date: '2026-11-25T10:05:00Z'
updated_date: '2026-11-25T12:24:00Z'
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
- [x] #1 nextjs renders all 8 component kinds (card, detail, table, list, hero, callout, picker, deck) with type-aware field rendering and zero frontend model slots
- [x] #2 Frontend prompts, component body verifier, and unused model-slot plumbing in internal/target/nextjs are removed cleanly
- [x] #3 board, library, neighbr, and booking specs and golden outputs are updated to : <kind> and theme: with zero frontend entries in philote.lock
- [x] #4 make check passes (golangci-lint, archtest, and full go test ./... including Next.js builds)
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Make : <kind> required on components and remove component look: across syntax, check, ir, and check tests. 2. Build Go view structs and templates for all 8 built-in component kinds (card, detail, table, list, hero, callout, picker, deck) with type-aware field rendering and zero frontend model slots. 3. Remove frontend prompt.go, component body verifier, and component lock nodes in build/steps.go. 4. Migrate board, library, neighbr, and booking specs, lockfiles, and cmd/philote/component_test.go; update goldens and run full make check.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented all 8 built-in component kind templates (card, detail, table, list, hero, callout, picker, deck) with type-aware field rendering in internal/target/nextjs/kind.go and templates/kind-*.tsx.tmpl, with zero frontend model slots. Made : <kind> required on components and removed component look:, frontend prompt.go, component body verifier, and component lock nodes in build/steps.go. Migrated board, library, neighbr, and booking specs and stripped all web/* pins from their philote.lock files. Verified with golangci-lint (0 issues), archtest, unit tests, and full tsc + eslint + next build on all 4 golden apps.
<!-- SECTION:FINAL_SUMMARY:END -->
