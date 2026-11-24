---
id: TASK-16
title: Design the built-in component kinds
status: To Do
assignee: []
created_date: '2026-11-24T11:46:00Z'
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
- [ ] #1 The design is recorded as a decision: the kinds, the syntax, rendering by field type, and theme choice
- [ ] #2 Every component in the neighbr, board, library and booking specs maps to a kind, or the gap is named
- [ ] #3 The user has accepted the new language concepts in it
<!-- AC:END -->
