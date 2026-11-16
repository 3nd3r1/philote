---
id: TASK-4
title: Make the one-part rebuild visible
status: To Do
assignee: []
created_date: '2026-11-16T12:19:00Z'
updated_date: '2026-11-16T12:20:00Z'
labels: []
milestone: m-5
dependencies: []
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When a spec changes, the person should see at a glance that only the touched part regenerated and everything else was reused. Today the build logs which nodes go to the model. Make philote build print, for every build, what was reused and what was regenerated, readable in two seconds. This is product output for every spec; the demo (m-5) films it on a neighbr edit that sends exactly one node to the model.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The chosen edit sends exactly one node to the model; every other generated file is byte-identical
- [ ] #2 philote build prints which nodes were reused and which were regenerated, readable at a glance
<!-- AC:END -->
