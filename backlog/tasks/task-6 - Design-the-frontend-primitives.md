---
id: TASK-6
title: Design the frontend primitives
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:33:00Z'
updated_date: '2026-11-17T10:39:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The frontend language has three fixed page kinds (list, record, form) and a fixed nav, so every philote app looks like the same admin panel, and a swipe deck or a landing page cannot be said at all (decision-3). The user wants apps as good as prompting Claude directly, still deterministic through the lock. The primitives sketched in conversation on 2026-11-17: a page shows components; a component takes one record or a list, can be interactive, and can call endpoints; nav (top, bottom, side) and landing copy live in the spec. These are new language concepts, so the design is the user's call.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A decision describes the primitives and shows neighbr's full spec rewritten in them, including a landing page, a bottom nav and a swipe deck
- [x] #2 The decision says how today's list, record and form pages are expressed in the primitives
- [x] #3 The user accepted the decision
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
decision-4 (proposed) written: five primitives (page load/shows, list params and uses, calls, says, nav) plus viewer, neighbr rewritten in them, mapping from list/record/form. Waiting on the user.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
decision-4 accepted by the user on 2026-11-17: pages load data and place components; components take lists, compose with uses, act with calls, get their words from says; the nav is in the spec; viewer picks who you are without login. decision-2 (model-written page layouts) is superseded. Built in four slices, each its own task.
<!-- SECTION:FINAL_SUMMARY:END -->
