---
id: TASK-4
title: Make the one-part rebuild visible
status: To Do
assignee: []
created_date: '2026-11-16T12:19:00Z'
labels: []
milestone: m-5
dependencies: []
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The second half of the video changes one thing in the neighbr spec and shows that only that part regenerates while the app looks exactly the same everywhere else. Pick the edit (for example adding a field to HorseCard's shows, or changing its look) so exactly one node goes to the model. Make the build output say it plainly, so a viewer sees it in two seconds: what was reused, what was regenerated.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The chosen edit sends exactly one node to the model; every other generated file is byte-identical
- [ ] #2 philote build prints which nodes were reused and which were regenerated, readable at a glance
<!-- AC:END -->
