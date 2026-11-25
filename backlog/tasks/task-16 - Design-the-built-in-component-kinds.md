---
id: TASK-16
title: Design the built-in component kinds
status: Done
assignee:
  - '@agy'
created_date: '2026-11-24T11:46:00Z'
updated_date: '2026-11-25T10:04:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 13000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-6: the frontend is built from templates only. A component becomes one of a fixed set of built-in kinds that the spec picks, rendered from its fields by type, and the look is one of a few fixed themes; the model writes no frontend code. This task designs that set before any code: which kinds (likely card, list, table, detail, form, hero, picker, deck), how the spec names a kind and passes it fields, words and actions, how fields render by type (dates, numbers, money, relations), how a theme is picked, and how the existing neighbr, board, library and booking specs map onto it. A component's kind is the one place its body comes from, so a later milestone can add a model-written kind.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The design is recorded as a decision: the kinds, the syntax, rendering by field type, and theme choice
- [x] #2 Every component in the neighbr, board, library and booking specs maps to a kind, or the gap is named
- [x] #3 The user has accepted the new language concepts in it
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Audit all components in board, library, neighbr, and booking specs. 2. Propose built-in component kinds, syntax, field-type rendering, and fixed themes to the user for sign-off. 3. Record the accepted design as a Backlog decision and close TASK-16.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Recorded the accepted built-in component kinds and fixed themes design in decision-7 (8 kinds: card, detail, table, list, hero, callout, picker, deck; : <kind> syntax; 4 fixed themes: plain, warm, editorial, slate; field-type rendering; all 21 components across board, library, neighbr, and booking mapped; money type gap named and deferred).
<!-- SECTION:FINAL_SUMMARY:END -->
