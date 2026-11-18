---
id: decision-5
title: The m-6 bar holds for apps besides neighbr
date: '2026-11-18T13:16:00Z'
status: accepted
---
## Context

Every frontend fix in m-6 so far was tuned on neighbr, and the m-6 bar in doc-1 is measured on neighbr alone. A compiler that only builds a good horse Tinder would pass. A first look on 2026-11-18 also showed the main gap is general: components are written one by one and share no visual language.

## Decision

The user chose this on 2026-11-18 ("even maybe multiple other apps"). The m-6 bar, not worse side by side than the same description prompted to Claude directly, must hold for neighbr and for two more apps written fresh: a booking app (services, times, booking a slot) and a recipe app (reading-heavy, a landing page, saving favourites). Each new app's spec and plain description are written before any compiler change made for it.

## Consequences

m-6 gets one task per new app after TASK-11. A fix found on one app has to keep the others no worse. The new specs become golden cases. doc-1's bar wording is the user's to update.
