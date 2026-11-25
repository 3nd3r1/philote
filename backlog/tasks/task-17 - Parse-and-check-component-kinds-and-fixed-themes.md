---
id: TASK-17
title: Parse and check component kinds and fixed themes
status: To Do
assignee: []
created_date: '2026-11-25T10:04:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-7 replaces free-form look strings with a fixed set of 8 built-in component kinds (: <kind>) and 4 fixed frontend themes (theme: <name>). The compiler frontend stages need to parse, validate, and lower these constructs with clear diagnostics before the nextjs target switches to template-only rendering.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 syntax parses component Name[(param)]: <kind> and frontend theme: <name>, and rejects removed look: lines with clear diagnostics
- [ ] #2 check validates all 8 kinds (card, detail, table, list, hero, callout, picker, deck) including param shape, allowed clauses, says counts, and picker/deck calls signatures, plus valid theme names (plain, warm, editorial, slate)
- [ ] #3 ir and lower carry Component.Kind and Part.Theme cleanly
- [ ] #4 Unit tests cover valid specs and error cases (message and position) for every kind and theme rule
- [ ] #5 docs/language.md documents : <kind>, the 8 kinds, and theme:
<!-- AC:END -->
