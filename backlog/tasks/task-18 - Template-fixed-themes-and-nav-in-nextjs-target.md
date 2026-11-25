---
id: TASK-18
title: Template fixed themes and nav in nextjs target
status: To Do
assignee: []
created_date: '2026-11-25T10:05:00Z'
labels: []
milestone: m-6
dependencies:
  - TASK-17
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-6 and decision-7 eliminate model calls for frontend themes and nav bars. The nextjs target needs 4 contrast-verified built-in CSS themes (plain, warm, editorial, slate) and built-in templates for top, bottom, and side nav placements with zero frontend model slots for theme or nav.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 nextjs emits contrast-verified light and dark CSS tokens for plain, warm, editorial, and slate without creating a theme model slot
- [ ] #2 nextjs renders top, bottom, and side nav bars from templates without creating a nav model slot
- [ ] #3 Theme contrast tests verify all 4 built-in themes pass readability checks
<!-- AC:END -->
