---
id: TASK-7
title: Pages load data and place components
status: To Do
assignee:
  - '@claude'
created_date: '2026-11-17T10:39:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
First slice of decision-4. Pages are fixed kinds today (data + layout list/record), with a template heading and grid, so every app looks the same and / is reserved. Replace that with load + shows: a page fetches named data and renders components in order, each a full-width section the model lays out. Components can take lists, render other components with uses, and get their copy from says. decision-2 page layouts go away.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 docs/language.md describes load, shows, list-of parameters, uses and says, and neighbr, board and library are rewritten in them
- [ ] #2 Each new word has a golden case and at least one error case (message and position)
- [ ] #3 A page renders its components in order with no template heading, grid or model-written page layout, and / is a normal route
- [ ] #4 A build sends a component body back to the model when a says line is missing from it
- [ ] #5 neighbr has a landing page at / and make check passes; --frozen rebuilds are byte-identical
<!-- AC:END -->
