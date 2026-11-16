---
id: decision-1
title: An app picks its look with a look line, written into a pinned theme
date: '2026-11-16T12:26:00Z'
status: accepted
---
## Context

Every app philote built looked the same: the base's near-black theme, a purple primary, and shadcn's base-sera style (square corners, tiny uppercase labels) hard-coded in the ui components and the component prompt. D-017 (docs/archive) accepted a shadcn preset code in the spec, but it was never built, needs network at build time, and means nothing when reviewing a spec.

## Decision

The user chose this on 2026-11-16, over built-in named themes and the preset code.

- A frontend can have a `look` line, like a component's: short, capped, optional.
- The model turns it into the app's theme: colors for light and dark, fonts from a fixed list, corner radius. The compiler checks the result and pins it in philote.lock like a component body.
- The base's ui components drop the hard-coded base-sera style, so the theme drives the whole look.
- Component prompts name only theme variables, never the theme's values, so changing the app's look rewrites only the theme.

## Consequences

D-017's preset code is replaced. A spec with no frontend look gets a good default theme, not the old black one. Changing the base ui style changes every component's prompt once, so every pinned body is rewritten once.
