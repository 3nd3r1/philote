---
id: TASK-10
title: Components call endpoints, and pages load from GET endpoints
status: Done
assignee:
  - '@claude'
created_date: '2026-11-18T08:22:00Z'
updated_date: '2026-11-18T08:54:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Fourth slice of decision-4. neighbr still swipes through a form with two dropdowns, and the swipe page shows every horse, including ones already swiped. A component that can act (calls) turns the swipe into a deck you drag, and GET endpoints with route inputs and list outputs let a page load what only steps can compute, like the candidates a horse has not swiped yet or its own matches.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/language.md describes calls, GET endpoints with route inputs, and list of outputs; each has a golden case and at least one error case (message and position)
- [x] #2 A component with calls is a client component handed a typed action; inputs bound to viewer are filled on the server, the body passes the rest, and a body that never calls the action or uses fetch goes back to the model
- [x] #3 A page can load a list from a GET endpoint with viewer as input, and sends a visitor without a viewer to the picker first
- [x] #4 neighbr swipes through a SwipeDeck on candidates and lists its own matches; checked in Chrome at laptop and phone width
- [x] #5 make check and make smoke pass; --frozen rebuilds are byte-identical
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Backend: endpoint output and step result take list of; a GET endpoint takes record inputs from its route ([horse] params, declared in input, sent as ids, missing is not_found) and no body; a custom route may sit under a resource path when it clashes with no resource route.
2. hono: route params read and found in the handler, list results returned; client calls take route params positionally.
3. Frontend load: load: <name> from <endpoint>(<input>: viewer, ...) on a GET endpoint with a list output; Page gets Endpoint/Args and a Viewer flag (shows or needs the viewer) that drives proxy, needViewer and the picker check.
4. calls: <endpoint>(<input>: viewer | <param> | <open>, ...) on a component, one custom non-GET endpoint, every input bound once. nextjs: a use-server action per call reads the viewer from the cookie and returns an error message; the component file becomes use client, its wrapper hands the body the action; prompt explains the action and allows hooks; Verify requires the call and rejects fetch. A page showing such a component needs the viewer.
5. neighbr: candidates and matches_of endpoints with steps; /swipe shows SwipeDeck(deck) from candidates(horse: viewer); /matches loads matches_of; the swipe form page goes. Docs, check error tests, goldens, real model bodies, smoke (GET endpoints, /swipe redirect, deck renders candidates), Chrome check.
Language calls: GET route inputs are declared in input like POST ones; SwipeDeck takes only the deck because the server action fills viewer, so components keep one parameter.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Built as planned (23cdbae). Calls I made: GET route inputs are declared in input like POST ones; SwipeDeck takes only the deck (the server action fills horse from the viewer cookie), so components keep one parameter; one calls per component, custom non-GET endpoints only; a page loads only lists from endpoints, every input the viewer. Step tests now copy api.philote without the read endpoints. Checked in Chrome: picker after a stale cookie, drag right to like Mister Ed made a match (viewer filled server-side), X passes, deck ends on its says line, /matches shows the match, 390px frames for /swipe and /matches, no console errors. make check, make smoke pass; two frozen builds identical.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Components can call a custom endpoint (calls: swipe(horse: viewer, target: deck, liked)): the component becomes a client component, a generated server action fills viewer inputs from the cookie and returns an error message, and Verify sends back a body that never calls its action or uses fetch. GET endpoints take records from their route ([horse], not_found when missing) and return list of outputs from steps; pages load lists from them with the viewer (load: deck from candidates(horse: viewer)), and such pages redirect to the picker. neighbr: candidates and matches_of endpoints, /swipe is a SwipeDeck over candidates, /matches shows the viewer own matches; the swipe form page is gone. Docs, check error cases, goldens, smoke and a Chrome run at laptop and phone width. Follow-up in m-6: the side-by-side check against prompting Claude directly. The deck look (next card barely peeks) is for the user to judge there.
<!-- SECTION:FINAL_SUMMARY:END -->
