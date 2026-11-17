---
id: TASK-9
title: Pick who you are with viewer
status: To Do
assignee:
  - '@claude'
created_date: '2026-11-17T10:39:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Third slice of decision-4. Swiping needs an acting horse, and auth is parked (Q-004). viewer: Horse lets the app remember which record you are in a cookie, with a picker instead of a login; loads and calls can use viewer, and a page that needs it sends you to the picker first.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 viewer and picks are documented, with golden and error cases
- [ ] #2 A page that loads with viewer redirects to the picker when none is chosen, and works after picking
- [ ] #3 neighbr has /pick and /me
<!-- AC:END -->
