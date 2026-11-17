---
id: TASK-9
title: Pick who you are with viewer
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:39:00Z'
updated_date: '2026-11-17T15:24:00Z'
labels: []
milestone: m-6
dependencies: []
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Third slice of decision-4. Swiping needs an acting horse, and auth is parked (Q-004). viewer: Horse lets the app remember which record you are in a cookie, with a picker instead of a login; loads and calls can use viewer, and a page that needs it sends you to the picker first.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 viewer and picks are documented, with golden and error cases
- [x] #2 A page that loads with viewer redirects to the picker when none is chosen, and works after picking
- [x] #3 neighbr has /pick and /me
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Scope: viewer, picks and C(viewer) sections. Loads that pass viewer to an endpoint (load: deck from candidates(horse: viewer)) need GET endpoints with inputs, so they come with the calls slice. Calls I made: picks needs then; exactly one picker page; the picker page cannot show the viewer; the viewer entity needs a resource with get. Redirect: generated src/proxy.ts sends a visitor without the cookie to the picker with a real 307 (with cacheComponents the page shell streams before a page can redirect); needViewer inside the streamed sections catches a stale cookie (client-side redirect). Pick is a no-JS form posting a bound server action; the cookie is httpOnly, lax, one year. Found while testing: TASK-8 nav used usePathname outside Suspense, which fails next build; fixed (6d80d7d) and TestFrontendsTypecheck now runs next build on every golden app. make smoke covers /me without a viewer (307 to /pick), /me with one, and posting a pick form.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
viewer: <Entity> in a frontend, picks: viewer then "<route>" on a list component, and C(viewer) as a page section. A generated proxy redirects pages that show the viewer to the picker when no viewer is chosen; picking sets a cookie through a server action. neighbr has /pick (HorsePicker + JoinHerd) and /me, with Me in the nav. Check errors, golden, Verify (Pick used), smoke checks; make check and make smoke pass. Also fixed a TASK-8 prerender bug in the nav and added a production next build to the golden frontend test.
<!-- SECTION:FINAL_SUMMARY:END -->
