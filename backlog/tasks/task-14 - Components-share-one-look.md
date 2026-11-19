---
id: TASK-14
title: Components share one look
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-18T14:03:00Z'
updated_date: '2026-11-19T09:15:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
In the TASK-11 side by side, neighbr prompted to Claude directly used one letter avatar on every card, the swipe deck, the profile, the matches and the form preview, so the app felt designed as one. Philote writes each component body on its own (D-019), so they share no visual language: the swipe card and the profile have no avatar, matches sit in small grid cells with wrapping names, the landing hero is plain text on the background, and forms are a bare template. Bodies must stay pinned per component, so changing one component still regenerates only that one.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every component that shows the same entity draws it the same way (for neighbr: the same avatar on HorseCard, the picker, the swipe card, the profile and the matches), without one component prompt depending on another
- [x] #2 A list page lays its items out to suit the component (wide rows for MatchCard, a grid for HorseCard), not always a 3-column grid
- [x] #3 Forms look like the rest of the app: an enum with a few values is chips, a string with a max shows a counter
- [x] #4 Changing one component in the spec still regenerates only that component (golden test)
- [ ] #5 make smoke passes and the user has looked at neighbr, the booking app and the recipe app
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Base Mark component: a record's picture, first letter of its display value on a tint from the theme's primary turned by a hash of the name. Every body prompt that shows a record by name offers Mark; built from the component alone (D-019).
2. List cards choose data-shape tile or row; the list grid becomes one centered column when a card is a row (CSS :has). The prompt gives laptop and phone widths.
3. Forms: an enum with up to 6 values is chips (radios, no JS); a string with a max gets a live Counter.
4. Body rules allow theme tints, gradients and decorative shapes, and say CardContent is already a column.
5. TestOneComponentChanges; rebuild bodies with the model; make smoke; check neighbr in Chrome at laptop and 390px.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-11-18: Mark (base) on every component that shows a record by name; list cards can mark data-row; forms get chips (enum <= 6 values) and a counter (string with max); TestOneComponentChanges; components can't take an imported name (Card, Mark...). Checked neighbr in Chrome: Mark consistent on picker, deck, HorseCard, profile, matches; swipe card now has a photo area; chips and counter work. Matches stayed grid cells and the landing stayed plain on a fresh redo, so the body rules now allow theme tints, gradients and decorative shapes. --redo takes a comma-separated list.

2026-11-19: committed ca172c8. Verified in Chrome (seeded 8 horses, 2 matches): the same Mark on picker, swipe card, HorseCard, profile and matches; MatchCard picks row and names fit at 390px; HorseCard, ProjectCard, AuthorCard pick tile; the landing is a full-bleed gradient hero; chips post their value (shetland) and the counter follows input (11/280). make check and make smoke pass. AC #5 waits on the booking and recipe specs (TASK-12, TASK-13) and on the user looking. Found on the way: the base CardContent is flex-col, so a body's flex items-center stacked; the prompt now says so. --redo takes a comma-separated list.
<!-- SECTION:NOTES:END -->
