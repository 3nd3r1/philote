---
id: TASK-15
title: A call shows what it caused
status: To Do
assignee:
  - '@claude'
created_date: '2026-11-18T14:04:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
In the TASK-11 side by side, the directly prompted neighbr shows an It is a match overlay when a like is mutual. Philote cannot: the swipe endpoint outputs the Swipe, match_if_mutual writes the Match without returning it, and the SwipeDeck body only gets an error message back from the call. Any app where an action has a consequence worth showing (a booking confirmed, a recipe saved to a list) needs the result of the call to reach the component.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 An endpoint can output what its steps caused (for neighbr: the swipe and the match, if one was made), and docs/language.md says how
- [ ] #2 A component that calls the endpoint gets the call result in its body, typed, so the model can show it (for neighbr: a match moment after a mutual like)
- [ ] #3 A golden case and an error case cover the new output form
- [ ] #4 make smoke passes, including a mutual like that returns the match
<!-- AC:END -->
