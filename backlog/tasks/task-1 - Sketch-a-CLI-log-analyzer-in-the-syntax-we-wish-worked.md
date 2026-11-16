---
id: TASK-1
title: Sketch a CLI log analyzer in the syntax we wish worked
status: To Do
assignee: []
created_date: '2026-11-16T12:11:00Z'
labels: []
milestone: m-0
dependencies: []
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Write a spec for a small CLI log analyzer (reads log files, groups and counts errors, prints a report) in the syntax we wish philote had. Ignore what the parser accepts today. Then run `philote check` on it. Every rejection is a concrete gap between the CRUD language and a general one. See doc-1 for why.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The spec is a .philote file in the repo, short enough to review in a minute
- [ ] #2 Each gap `philote check` finds is listed in this task's notes, one line each
- [ ] #3 The notes say whether the spec is shorter than the code it replaces, and by roughly how much
<!-- AC:END -->
