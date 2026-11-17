---
id: TASK-8
title: The nav is described in the spec
status: To Do
assignee:
  - '@claude'
created_date: '2026-11-17T10:39:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Second slice of decision-4. The navbar is a fixed base file, the same top bar in every app. The spec says where the nav sits (top, bottom, side) and its items; the template passes items and the active one, and the model draws it from the app look, pinned like a component.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 nav: <place>, <route> as <label>, ... is documented, with golden and error cases
- [ ] #2 The nav is model-written and pinned; changing an item rewrites only the nav
- [ ] #3 neighbr has a bottom nav that works at phone width
<!-- AC:END -->
