+++
id = "D-006"
title = "Strict defaults from the prototype"
status = "proposed"
summary = "List pages must say what shows when empty, lookup pages must handle not found, `id` is reserved, raw markdown can't be shown as text, UI copy is capped at 200 chars."
created = 2026-11-09
answers = ["Q-002"]
related = ["P-001"]
+++

## Decision

- A list page must declare `empty`. A lookup page must declare `not_found`.
- `id` is reserved; every entity gets one.
- Showing a markdown field as plain text is an error: use `as markdown` or `excerpt(...)`.
- Lookups must go through a unique field (slugs count as unique).
- Natural-language UI copy is capped at 200 characters.

## Why

Each of these is a place where the model would otherwise guess. Making the author decide is the point of P-001.
