---
id: TASK-12
title: 'Side by side: a booking app from philote vs prompted to Claude directly'
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-18T13:16:00Z'
updated_date: '2026-11-19T09:56:00Z'
labels: []
milestone: m-6
dependencies:
  - TASK-11
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-5: the m-6 bar must hold beyond neighbr, so fixes tuned on one app cannot pass it. A booking app stresses different primitives than a swipe deck: services, opening times and booking a free slot as the viewer; forms, dates and a step that computes free slots. The spec and the plain description are written first, before any compiler change for this app, so the compiler is tested rather than fitted.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A plain description and a .philote spec for the booking app are committed before any compiler change made for it, and the spec is a golden case
- [ ] #2 The same description prompted to Claude directly runs locally next to philote's build, compared page by page at laptop and phone width, with every gap listed in this task's notes
- [ ] #3 Each gap is fixed in a way that improves any spec, and neighbr and the other apps are no worse after it
- [ ] #4 The user has looked at both apps and agrees philote's is not worse
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
2026-11-19: description and spec in cmd/philote/testdata/booking (chair, a hair salon). Writing the spec against the current compiler found language gaps, worked around in the spec for now:
L1 A datetime can't be a display field, so a Booking can't show its Slot's time (BookingCard shows booked_at and note only, which is worse than the description).
L2 delete takes no label: the description says Cancel, the spec gets the default.
L3 A GET route can't be /clients/[client]/bookings, the has-many read route; used /clients/[client]/upcoming.
L4 Two hops can't be shown: a booking can't show its slot's service.

2026-11-19 philote build (49340de), seeded 4 services, 27 slots, 2 clients; checked in Chrome at 1568px and in 390px iframes. Works end to end: pick a client, tap a slot, it's booked (check mark and booked time), My bookings lists it. Good: hero, theme (stone and olive, serif headings), nav, forms with counters.
Gaps found without the direct app yet:
P1 My bookings shows only Booked at: no slot time, no service (L1, L4). The page heading says Bookings, not My bookings.
P2 Cancel is a Delete button (L2).
P3 Prices are a bare 55: no currency (no money type, L5).
P4 The slot picker is one card per slot with the full date, not grouped by day; long on a phone.
P5 Services get letter avatars (C for Cut and Color), which reads oddly for services.
Direct app not run yet: launching a nested claude -p with bypassed permissions was blocked; waiting on the user.
<!-- SECTION:NOTES:END -->
