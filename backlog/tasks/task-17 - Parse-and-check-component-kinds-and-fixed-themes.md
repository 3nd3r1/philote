---
id: TASK-17
title: Parse and check component kinds and fixed themes
status: Done
assignee:
  - '@agy'
created_date: '2026-11-25T10:04:00Z'
updated_date: '2026-11-25T11:52:00Z'
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
- [x] #1 syntax parses component Name[(param)]: <kind> and frontend theme: <name>, and rejects removed look: lines with clear diagnostics
- [x] #2 check validates all 8 kinds (card, detail, table, list, hero, callout, picker, deck) including param shape, allowed clauses, says counts, and picker/deck calls signatures, plus valid theme names (plain, warm, editorial, slate)
- [x] #3 ir and lower carry Component.Kind and Part.Theme cleanly
- [x] #4 Unit tests cover valid specs and error cases (message and position) for every kind and theme rule
- [x] #5 docs/language.md documents : <kind>, the 8 kinds, and theme:
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add Component.Kind and Part.Theme to syntax AST, parser, and ir. 2. Add check validation for all 8 built-in component kinds (card, detail, table, list, hero, callout, picker, deck) and 4 fixed themes (plain, warm, editorial, slate). 3. Update lower.Tree to print component kind and theme when set. 4. Add comprehensive unit tests in internal/check covering valid specs and error cases (message + position) for every kind and theme rule. 5. Update docs/language.md. Note: keep look: and omitted : <kind> accepted temporarily so existing golden tests stay green until TASK-18 and TASK-19 switch nextjs and migrate golden specs.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Kept look: and omitted : <kind> temporarily accepted for untyped legacy specs so existing golden tests remain green until TASK-18 (removes frontend look:) and TASK-19 (removes component look: and requires : <kind>). look: is already rejected on typed components and alongside theme:.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added : <kind> component syntax and theme: <name> frontend syntax across syntax, check, ir, and lower; implemented full check validation for all 8 built-in component kinds (card, detail, table, list, hero, callout, picker, deck) and 4 fixed themes (plain, warm, editorial, slate); added 22 unit test cases in internal/check/kind_test.go; updated docs/language.md.
<!-- SECTION:FINAL_SUMMARY:END -->
