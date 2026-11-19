---
id: TASK-15
title: A call shows what it caused
status: Done
assignee:
  - '@claude'
created_date: '2026-11-18T14:04:00Z'
updated_date: '2026-11-19T09:46:00Z'
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
- [x] #1 An endpoint can output what its steps caused (for neighbr: the swipe and the match, if one was made), and docs/language.md says how
- [x] #2 A component that calls the endpoint gets the call result in its body, typed, so the model can show it (for neighbr: a match moment after a mutual like)
- [x] #3 A golden case and an error case cover the new output form
- [x] #4 make smoke passes, including a mutual like that returns the match
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Language: a step result can be optional (`-> match: Match optional`, the step returns the record or null). An endpoint output can be a record of named values: `output: { swipe: Swipe, match: Match optional }`, each a step result or input of that exact type.
2. check: optional only on a single record result; an optional result can't feed a step parameter; output record fields must name a value in scope with the same type. Error cases with positions.
3. IR: Value.Optional; Endpoint.Outputs for a record output.
4. hono: steps return X | null; the handler returns the record { swipe, match }.
5. nextjs: api.ts types the record; the call action resolves to { ok: true, value } or { ok: false, error }; the body prompt spells the value's type so the model can show what the call caused.
6. neighbr: swipe outputs { swipe, match }, SwipeDeck says It's a match and shows a match moment. docs/language.md, golden + error cases, smoke checks a mutual like returns the match.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-11-19 (c8b1247): output: { swipe: Swipe, match: Match optional }; a step returns one record optional (X | null); an optional result can't feed a step or a plain output. The call action resolves to { ok: true, value } | { ok: false, error }, typed as <Endpoint>Result and spelled field by field in the body prompt. Verified: check tests TestNamedOutputs and TestNamedOutputErrors (9 cases with positions); neighbr golden; make smoke checks the mutual like returns the match over curl and through the client (a one-way like returns match null); in Chrome, liking a horse that liked you opens It's a match! with its Mark, and Keep swiping shows the next card.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Endpoints can return what their steps made by name, and a calling component gets that result typed, so neighbr's deck shows a match moment after a mutual like. Language: `output: { name: Type, ... }` and `-> name: Entity optional`, documented in docs/language.md. Hono returns the record; Next.js call actions resolve to { ok, value } or { ok: false, error }. make check and make smoke pass; checked in Chrome. The SwipeDeck prompt now also depends on the Swipe and Match entities, since it shows them.
<!-- SECTION:FINAL_SUMMARY:END -->
