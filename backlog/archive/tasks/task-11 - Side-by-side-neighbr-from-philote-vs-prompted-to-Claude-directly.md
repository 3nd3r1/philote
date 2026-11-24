---
id: TASK-11
title: 'Side by side: neighbr from philote vs prompted to Claude directly'
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-18T11:18:00Z'
updated_date: '2026-11-18T14:17:00Z'
labels: []
milestone: m-6
dependencies:
  - TASK-14
  - TASK-15
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The bar for m-6 (doc-1) is that neighbr built by philote is not worse than the same app prompted to Claude directly, swipe deck and landing page included. All four slices of decision-4 are built (TASK-7..10), but nobody has compared yet. The comparison decides whether m-6 is done or needs another slice. A first look on 2026-11-18 found: the swipe card has no avatar while HorseCard has one, the deck card is mostly empty, a mutual like gives no match moment, and /matches is one small card in a corner with no heading. Components are written one by one, so they share no visual language.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The same description of neighbr is prompted to Claude directly and that app runs locally next to philote's build
- [x] #2 Both are compared page by page at laptop and phone width, with screenshots, and every place philote's app is worse is listed in this task's notes
- [x] #3 Each gap is fixed in a way that improves any spec (no neighbr-only fixes), or becomes its own m-6 task
- [ ] #4 The user has looked at both apps and agrees philote's is not worse
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Write neighbr as plain prose (same pages, data, looks) and prompt Claude directly (claude -p, claude-opus-5-5, same stack: Hono + Next.js) in an empty folder.
2. Build philote neighbr from the testdata spec (make smoke output), run both against the same seed horses.
3. Screenshot both page by page at laptop and 390px phone width; list every gap in the notes.
4. Fix each gap in the compiler/templates/prompts so any spec improves; split big ones into m-6 tasks.
5. Show the user both apps.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-11-18 side by side. Direct: the prompt in plain prose (same data, pages, looks, nav), claude -p with claude-opus-5-5, Hono + Next.js, 74 turns, 7.5 min, about 2 USD. Both seeded with the same 8 horses and 2 matches for Sea Biscuit; compared at 1568px and in a 390px iframe. darkreader-lock added to the direct layout so Dark Reader does not recolor it (philote base has it).

Where philote is worse:
G1 Theme mode: philote follows the system dark mode, so for a dark-mode visitor coral and cream shows as dark brown. Direct is always cream. In light mode philote is close.
G2 Page headings: direct has Horses and Matches headings; philote list pages start straight into cards.
G3 No shared visual language: direct has one letter avatar (LetterArt) on every card, the deck, the profile, the matches and the form preview. Philote has it on HorseCard and the picker only; the swipe card and the profile have none.
G4 Swipe card: philote is a title, an AGE label and a bio on a mostly empty card; direct is a photo-style avatar area with name, age, badge and bio under it, like Tinder.
G5 Matches: philote puts each match in a small grid cell with names wrapping; direct uses wide rows with both avatars.
G6 No match moment: direct shows an It is a match overlay after a mutual like; philote only removes the card. The swipe endpoint returns the Swipe, so the client cannot know a match happened.
G7 Landing hero: direct is a full-bleed coral gradient with soft shapes; philote is centered text on the page background.
G8 Forms: philote is a plain template form (native select, small Save, New Horse); direct has breed chips, a live avatar preview and a character counter.
G9 Small: direct confirms delete inline (philote uses window.confirm), says Swiping as Sea Biscuit, has arrow-key swiping and LIKE/NOPE stamps.

Same or better in philote: phone layout, the picker, HorseCard, the nav, byte-stable output and pinned bodies.

Fixed here, for any spec: G1 the theme sets --mode (light, dark or system) from the look and the app opens in it (neighbr: light, cream); G2 a page that opens with a list shows its title as a heading; G9 delete asks inline with Cancel instead of window.confirm. make check and make smoke pass. Split out: G3 G4 G5 G7 G8 into TASK-14 (components share one look); G6 into TASK-15 (a call shows what it caused). Not taken: the direct app invented copy (Meet the herd, Swiping as, 8 horses in the herd); philote words come only from the spec.
<!-- SECTION:NOTES:END -->
