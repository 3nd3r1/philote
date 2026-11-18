---
id: TASK-13
title: 'Side by side: a recipe app from philote vs prompted to Claude directly'
status: To Do
assignee:
  - '@claude'
created_date: '2026-11-18T13:16:00Z'
labels: []
milestone: m-6
dependencies:
  - TASK-11
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-5: the m-6 bar must hold beyond neighbr, so fixes tuned on one app cannot pass it. A recipe app stresses different primitives than a swipe deck: reading: a landing page, long markdown recipes, browsing by category, saving favourites as the viewer. The spec and the plain description are written first, before any compiler change for this app, so the compiler is tested rather than fitted.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A plain description and a .philote spec for the recipe app are committed before any compiler change made for it, and the spec is a golden case
- [ ] #2 The same description prompted to Claude directly runs locally next to philote's build, compared page by page at laptop and phone width, with every gap listed in this task's notes
- [ ] #3 Each gap is fixed in a way that improves any spec, and neighbr and the other apps are no worse after it
- [ ] #4 The user has looked at both apps and agrees philote's is not worse
<!-- AC:END -->
