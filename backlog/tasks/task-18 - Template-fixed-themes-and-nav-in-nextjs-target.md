---
id: TASK-18
title: Template fixed themes and nav in nextjs target
status: Done
assignee:
  - '@agy'
created_date: '2026-11-25T10:05:00Z'
updated_date: '2026-11-25T12:04:00Z'
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
- [x] #1 nextjs emits contrast-verified light and dark CSS tokens for plain, warm, editorial, and slate without creating a theme model slot
- [x] #2 nextjs renders top, bottom, and side nav bars from templates without creating a nav model slot
- [x] #3 Theme contrast tests verify all 4 built-in themes pass readability checks
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add 4 contrast-verified fixed themes (plain, warm, editorial, slate) in internal/target/nextjs/theme.go and emit globals.css with zero theme model slots. 2. Template top, bottom, and side nav bars in internal/target/nextjs/nav.go and templates/main-navbar.tsx.tmpl with zero nav model slots. 3. Remove frontend look: (keep component look: until TASK-19), remove theme/nav slots from build/verify, and add unit tests verifying all 4 built-in themes pass checkTheme contrast checks. 4. Update cmd/philote/theme_test.go, nav_test.go, and neighbr/booking specs (theme: warm / theme: editorial) + goldens.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added 4 contrast-verified built-in CSS themes (plain, warm, editorial, slate) in internal/target/nextjs/theme.go and templated top/bottom/side nav bars in main-navbar.tsx.tmpl with zero theme or nav model slots. Removed frontend look: from syntax/check/ir/lower/build. Verified all 4 themes with checkTheme contrast tests, updated TestFixedTheme and TestNavFromSpec (0 model calls, tsc+eslint clean), and updated neighbr/booking specs and goldens.
<!-- SECTION:FINAL_SUMMARY:END -->
