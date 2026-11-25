---
id: decision-8
title: Demo app is a planet-booking app (Airbnb for planets) instead of neighbr
date: '2026-11-25T11:41:00Z'
status: accepted
---
## Context

After decision-6 (templated frontend, model writes backend logic only), `neighbr` puts the spotlight on consumer dating-app polish and swipe physics — philote's weakest angle. A booking/workflow app shows off philote's real strengths (strict data models, relations, bounded pinned backend steps, and clean deterministic UI), while still needing a fun, concrete premise that hooks viewers in the first seconds of the m-5 demo video.

## Decision

The user chose this on 2026-11-25: the m-5 demo app is a **planet-booking app ("Airbnb for planets")** instead of `neighbr`.

- **Domain**: travelers browse planets (`Planet`), pick a stay/launch window (`Window`), and book stays (`Booking`) with backend double-booking enforcement and per-traveler views.
- **UI**: built strictly from the decision-7 built-in kinds (`hero`, `card`, `detail`, `picker`, `callout`, `form`) and a fixed theme (`editorial` or `warm`) — zero frontend model calls.
- **Demo arc**: first 30s compiles a short spec into a working full-stack planet-booking app; last 30s makes a one-line spec edit (e.g. adding a field or tweaking a backend step's `intent`) and shows deterministic, surgical regeneration.

## Consequences

- `neighbr` stays as a test spec (keeping `deck` covered in m-6 so nothing regresses), while the m-5 demo spec will be the planet-booking app.
- `doc-1`'s m-5 demo wording ("filmed on neighbr") is the user's to update when ready.

