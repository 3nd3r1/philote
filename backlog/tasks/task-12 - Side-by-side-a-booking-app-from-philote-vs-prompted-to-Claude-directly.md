---
id: TASK-12
title: 'Side by side: a booking app from philote vs prompted to Claude directly'
status: To Do
assignee:
  - '@claude'
created_date: '2026-11-18T13:16:00Z'
labels: []
milestone: m-6
dependencies:
  - TASK-11
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-5: the m-6 bar must hold beyond neighbr, so fixes tuned on one app cannot pass it. A booking app stresses different primitives than a swipe deck: services, opening times and booking a free slot as the viewer; forms, dates and a step that computes free slots. The spec and the plain description are written first, before any compiler change for this app, so the compiler is tested rather than fitted.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A plain description and a .philote spec for the booking app are committed before any compiler change made for it, and the spec is a golden case
- [ ] #2 The same description prompted to Claude directly runs locally next to philote's build, compared page by page at laptop and phone width, with every gap listed in this task's notes
- [ ] #3 Each gap is fixed in a way that improves any spec, and neighbr and the other apps are no worse after it
- [ ] #4 The user has looked at both apps and agrees philote's is not worse
<!-- AC:END -->
