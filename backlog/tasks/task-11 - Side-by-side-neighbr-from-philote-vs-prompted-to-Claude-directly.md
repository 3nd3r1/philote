---
id: TASK-11
title: 'Side by side: neighbr from philote vs prompted to Claude directly'
status: To Do
assignee:
  - '@claude'
created_date: '2026-11-18T11:18:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The bar for m-6 (doc-1) is that neighbr built by philote is not worse than the same app prompted to Claude directly, swipe deck and landing page included. All four slices of decision-4 are built (TASK-7..10), but nobody has compared yet. The comparison decides whether m-6 is done or needs another slice. A first look on 2026-11-18 found: the swipe card has no avatar while HorseCard has one, the deck card is mostly empty, a mutual like gives no match moment, and /matches is one small card in a corner with no heading. Components are written one by one, so they share no visual language.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The same description of neighbr is prompted to Claude directly and that app runs locally next to philote's build
- [ ] #2 Both are compared page by page at laptop and phone width, with screenshots, and every place philote's app is worse is listed in this task's notes
- [ ] #3 Each gap is fixed in a way that improves any spec (no neighbr-only fixes), or becomes its own m-6 task
- [ ] #4 The user has looked at both apps and agrees philote's is not worse
<!-- AC:END -->
