---
id: TASK-3
title: Apps philote builds look great by default
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-16T12:19:00Z'
updated_date: '2026-11-16T12:49:00Z'
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

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Found (neighbr built --frozen, 2026-11-16): every app gets the same look from the base: near-black background, grey cards, a purple primary, and shadcn's base-sera style (square corners, tiny uppercase tracking-wide labels and buttons). The component prompt hard-codes that style. The spec can't choose a look (D-017's theme preset was never built). / is an empty placeholder. Detail-page actions float under the card.

1. Direction call to the user: how an app picks its look.
2. / redirects to the first nav page when the spec has no / page.
3. Detail and form pages: actions and fields sit inside the layout; labels readable.
4. Component prompt describes the theme's style, not base-sera's.
5. Check neighbr, board and library on a laptop and phone-size window; --frozen stays byte-identical.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
decision-1 (accepted by the user): an app picks its look with a frontend look line, written by the model into a pinned theme.

2026-11-16 progress (committed):
- Built decision-1: frontend `look:` line; the model writes web/theme (colors light+dark, font from a fixed list of 11, radius), checked in Go (format, font list, WCAG contrast) and pinned in philote.lock. No look = default theme, no model call. Changing the look rewrites only the theme (TestThemeFromLook).
- Base: shadcn base-sera -> base-vega (normal case, radius from theme); globals.css is now a template with a fenced theme; all fonts loaded with preload false; light/dark follows the system, with a toggle; navbar shows the app name, links get their own row on phones; / redirects to the first nav page.
- Component prompt: no base-sera style; "show each value once".
- neighbr has a look line (coral and cream, Nunito + Fraunces). Bodies regenerated for neighbr, board and library; goldens updated; make check passes; next build passes.
Checked in Chrome: neighbr dark, matches light, phone width 390px; board default theme dark.
Left before Done: look at library; check light mode with Dark Reader off (the user's Chrome runs Dark Reader, which recolors light pages and causes the hydration warnings); neighbr's HorseCard hero is empty by design of its look line, judge whether the spec's look should change.
<!-- SECTION:NOTES:END -->
