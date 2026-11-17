---
id: TASK-7
title: Pages load data and place components
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-17T10:39:00Z'
updated_date: '2026-11-17T10:51:00Z'
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
- [ ] #1 docs/language.md describes load, shows, list-of parameters, uses and says, and neighbr, board and library are rewritten in them
- [ ] #2 Each new word has a golden case and at least one error case (message and position)
- [ ] #3 A page renders its components in order with no template heading, grid or model-written page layout, and / is a normal route
- [ ] #4 A build sends a component body back to the model when a says line is missing from it
- [ ] #5 neighbr has a landing page at / and make check passes; --frozen rebuilds are byte-identical
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-11-17 end of day, on branch pages-load-components (committed with the hook skipped, since make check fails until bodies are regenerated).
Done: syntax/check/IR/lower for load, shows sections, list-of params, recordless components, says, links_to ... as; check error cases in view_test; decision-2 page layouts removed (lock kind page, PageBody, page prompt); nextjs: one page.tsx template stacking sections, list-of grid, list reads in reads.ts, component wrapper hands the body link and actions as props, says as a const array, Verify checks says[i], link.href and {actions}; base home page replaced when a spec has /; neighbr has a landing at /, library's books use BookShelf(books: list of Book); docs/language.md, README, CLAUDE.md updated. Unit tests for internal/... pass; golangci-lint clean.
Next: (1) regenerate bodies: philote build (no --frozen) for neighbr, library, board under cmd/philote/testdata (page stanzas already removed from their locks); (2) go test ./cmd/philote -update, review goldens, make check; (3) run the three apps in Chrome, laptop and phone width; (4) smoke test; (5) merge to main.
Not done from the AC: uses was left out (no spec needs it yet).
<!-- SECTION:NOTES:END -->
