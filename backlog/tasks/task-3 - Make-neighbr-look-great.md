---
id: TASK-3
title: Apps philote builds look great by default
status: Done
assignee:
  - '@claude'
created_date: '2026-11-16T12:19:00Z'
updated_date: '2026-11-17T08:41:00Z'
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
- [x] #1 neighbr, board and library each look like a polished app in screenshots on a laptop and a phone-size window
- [x] #2 Nothing in the compiler or base is specific to one spec
- [x] #3 Rebuilding with --frozen gives byte-identical output
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

2026-11-17: library checked: clean in dark and light, but card and view actions float under the card on every spec (plan item 3 not done), view bodies wrap their Card in their own page padding, and some bodies override CardHeader so CardAction sits mid-row. Fix: the component template frames body and actions as one card with an actions bar; the prompt asks for a bare Card root with no layout of its own and CardHeader kept as is; Verify sends back a body whose root is not a Card. Bodies regenerated.

Verified 2026-11-17 in Chrome after the fix: library, neighbr and board on a laptop width and in a 390px frame (the window manager pins the window at full width, so phone width was checked in a 390px iframe). Every record with actions is one card with an actions bar at its bottom; CardAction sits top right. Library light mode looked clean before Dark Reader kicked in; later light screenshots are Dark Reader's recoloring. neighbr's HorseCard look line now asks for the name's initial on a primary tint where a photo would go, instead of an empty hero; rebuilding asked the model for HorseCard only. Two --frozen builds of each example are byte-identical and match the goldens; make check passes.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Apps philote builds now look like polished apps by default, for any spec. An app picks its look with a frontend look line, which the model writes into a pinned, Go-checked theme (decision-1); with no look line the app gets a neutral default theme. The base moved to shadcn base-vega, follows the system light/dark setting with a toggle, and / redirects to the first page. Component actions now sit in a bar joined to the body's card instead of floating under it. The prompt asks for a bare Card root with no layout of its own, and Verify sends back any body that doesn't return one. Checked on library, neighbr and board at laptop and phone widths; --frozen builds are byte-identical. Follow-ups: none needed for m-5. Light mode under the user's Dark Reader will always look recolored, and that isn't an app bug.
<!-- SECTION:FINAL_SUMMARY:END -->
