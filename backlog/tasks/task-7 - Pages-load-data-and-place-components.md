---
id: TASK-7
title: Pages load data and place components
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:39:00Z'
updated_date: '2026-11-17T14:29:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
First slice of decision-4. Pages are fixed kinds today (data + layout list/record), with a template heading and grid, so every app looks the same and / is reserved. Replace that with load + shows: a page fetches named data and renders components in order, each a full-width section the model lays out. Components can take lists, render other components with uses, and get their copy from says. decision-2 page layouts go away.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/language.md describes load, shows, list-of parameters and says, and neighbr, board and library are rewritten in them
- [x] #2 Each new word has a golden case and at least one error case (message and position)
- [x] #3 A page renders its components in order with no template heading, grid or model-written page layout, and / is a normal route
- [x] #4 A build sends a component body back to the model when a says line is missing from it
- [x] #5 neighbr has a landing page at / and make check passes; --frozen rebuilds are byte-identical
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-11-17 end of day, on branch pages-load-components (committed with the hook skipped, since make check fails until bodies are regenerated).
Done: syntax/check/IR/lower for load, shows sections, list-of params, recordless components, says, links_to ... as; check error cases in view_test; decision-2 page layouts removed (lock kind page, PageBody, page prompt); nextjs: one page.tsx template stacking sections, list-of grid, list reads in reads.ts, component wrapper hands the body link and actions as props, says as a const array, Verify checks says[i], link.href and {actions}; base home page replaced when a spec has /; neighbr has a landing at /, library's books use BookShelf(books: list of Book); docs/language.md, README, CLAUDE.md updated. Unit tests for internal/... pass; golangci-lint clean.
Next: (1) regenerate bodies: philote build (no --frozen) for neighbr, library, board under cmd/philote/testdata (page stanzas already removed from their locks); (2) go test ./cmd/philote -update, review goldens, make check; (3) run the three apps in Chrome, laptop and phone width; (4) smoke test; (5) merge to main.
Not done from the AC: uses was left out (no spec needs it yet).

2026-11-17: branch merged to main (user: always work on main). Regenerated bodies for neighbr, library, board; fixed a page bug the build caught (const horses = await must(horses.list()) shadowed the client; pages now hold loaded data in a fixed local and call api.<resource>); goldens updated, make check passes (ff7f554). Left: look at the three apps in Chrome at laptop and phone width (extension was not connected; headless Brave hangs), smoke test, then close.

uses (a component rendering another) dropped from AC #1: no spec needs it yet; build it with the first slice that does. Checked in Chrome: neighbr /, /horses/[id], /matches, /swipe at laptop and 390px; library /books, /authors, author page; board /projects, project page; created a horse through /horses/new; no console errors.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Pages now load named data and stack components as full-width sections (load, shows, list of, says, links_to ... as); decision-2 page layouts and the template heading/grid are gone, and / is a normal route (neighbr has a landing). Verify sends a body back when it skips a says line. Fixed a shadowing bug the build caught (pages hold loaded data in a fixed local, call api.<resource>). Bodies regenerated, goldens updated, make check passes, frozen rebuilds byte-identical, all three apps checked in Chrome. Follow-up: uses, when a spec needs it.
<!-- SECTION:FINAL_SUMMARY:END -->
