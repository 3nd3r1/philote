---
id: TASK-14
title: Components share one look
status: To Do
assignee:
  - '@claude'
created_date: '2026-11-18T14:03:00Z'
updated_date: '2026-11-18T14:03:00Z'
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
- [ ] #1 Every component that shows the same entity draws it the same way (for neighbr: the same avatar on HorseCard, the picker, the swipe card, the profile and the matches), without one component prompt depending on another
- [ ] #2 A list page lays its items out to suit the component (wide rows for MatchCard, a grid for HorseCard), not always a 3-column grid
- [ ] #3 Forms look like the rest of the app: an enum with a few values is chips, a string with a max shows a counter
- [ ] #4 Changing one component in the spec still regenerates only that component (golden test)
- [ ] #5 make smoke passes and the user has looked at neighbr, the booking app and the recipe app
<!-- AC:END -->
