---
id: TASK-8
title: The nav is described in the spec
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:39:00Z'
updated_date: '2026-11-17T14:45:00Z'
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
- [x] #1 nav: <place>, <route> as <label>, ... is documented, with golden and error cases
- [x] #2 The nav is model-written and pinned; changing an item rewrites only the nav
- [x] #3 neighbr has a bottom nav that works at phone width
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Calls I made: bottom and top mean that place on every screen size; side is a full-height column, icons only on phones. The template owns placement (a sticky wrapper; layout.tsx puts a bottom nav after the page and a side nav beside it) and the active item (longest matching href, / only itself); the model draws only what is inside. Item text defaults to the page title, cap 24. Without a nav line the template draws the old top bar over the old items, now with the active item marked, and no model call. Verify checks the nav uses items.map, active, home.href, home.label and renders {toggle} once; body checks now skip the signature. neighbr: nav: bottom, /swipe, /horses, /matches, /horses/new as Add horse; checked in Chrome at laptop and 390px. Side nav is covered by TestNavFromSpec but not looked at in a browser.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
nav: <top|bottom|side>, "<route>" [as "<label>"], ... in a frontend. The model draws the nav from the look, pinned as nav web/nav; the template keeps it in place and finds the active item. Check errors for place, missing or by-id pages, duplicates, empty or long labels, two navs, nav in a backend. TestNavFromSpec covers the retry loop and that changing an item asks only for the nav. neighbr has a bottom tab bar. make check passes.
<!-- SECTION:FINAL_SUMMARY:END -->
