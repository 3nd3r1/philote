---
id: TASK-3
title: Apps philote builds look great by default
status: To Do
assignee: []
created_date: '2026-11-16T12:19:00Z'
updated_date: '2026-11-16T12:20:00Z'
labels: []
milestone: m-5
dependencies: []
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Every frontend philote builds looks like the plain dark default of the Next.js base. For m-5 (doc-1) the product, not just neighbr, has to look like a polished consumer app. Decide where visual design lives (base theme, the theme tokens accepted in D-017 in docs/archive, or both) and build it for every spec. Nothing special-cased for neighbr. Neighbr, board and library are the checks.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 neighbr, board and library each look like a polished app in screenshots on a laptop and a phone-size window
- [ ] #2 Nothing in the compiler or base is specific to one spec
- [ ] #3 Rebuilding with --frozen gives byte-identical output
<!-- AC:END -->
