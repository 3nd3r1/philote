---
id: decision-7
title: Built-in component kinds and fixed themes
date: '2026-11-25T10:03:00Z'
status: accepted
---
## Context

decision-6 moves philote's frontend to templates only so every build is deterministic and requires zero frontend model calls, while keeping a component's kind as the single dispatch point so a later milestone can add a model-written kind without reworking pages. TASK-16 audited all 21 components across the existing `board`, `library`, `neighbr`, and `booking` specs to define the fixed set of built-in kinds, syntax, field-type rendering, and fixed themes.

## Decision

The user accepted this design on 2026-11-25.

1. **Syntax**:
   - Component kind goes in the component header with `: <kind>` (matching `backend api: hono` and `frontend web: nextjs`), e.g. `component HorseCard(horse: Horse): card { ... }` and `component Landing: hero { ... }`.
   - `look:` is removed from components and from `frontend`.
   - `frontend` adds optional `theme: <name>` (defaults to `plain`).
   - `nav: top | bottom | side, ...` keeps its existing syntax and is rendered by built-in templates (no model calls).
   - Forms remain on `page "/..." { form: ... }` (already templated).

2. **Built-in component kinds (8)**:
   - `card` (1 record): `shows`, `links_to`, `actions`. Compact card (avatar if display is string, title, typed metadata); used in `list of C(items)` grids or standalone.
   - `detail` (1 record): `shows`, `links_to ... as`, `actions`. Full-width record view with large header, labeled metadata grid, long text (`max > 140` / `markdown`) prose block, `as chips` section, and action bar.
   - `table` (`list of`): `shows`. Multi-column table, one column per shown field with type-aware formatting.
   - `list` (`list of`): `shows`. Compact stacked vertical rows.
   - `hero` (no record): `says` (1-2), `links_to ... as`. Full-width landing section (`says[0]` headline, optional `says[1]` subtitle, optional primary CTA).
   - `callout` (no record): `says` (1), `links_to ... as`. Compact centered line with a quiet secondary button.
   - `picker` (`list of`): `shows`, `says` (0-1), and either `picks: viewer then "<route>"` or `calls: <endpoint>` (taking only viewer + tapped record). Selectable tile grid.
   - `deck` (`list of`): `shows`, `says` (1-3), and `calls: <endpoint>` (taking viewer + current record + one `bool`). Swipe card stack with drag and pass/like buttons; `says[0]` empty state, `says[1..2]` match overlay when endpoint returns an optional record.

3. **Rendering by field type**:
   - `string`: display field as title (with deterministic letter avatar in `card`/`picker`/`detail`); `max > 140` as body paragraph; shorter strings as labeled values.
   - `slug`: muted monospace pill.
   - `email`: `mailto:` link.
   - `markdown`: formatted prose block.
   - `int`: locale-formatted `tabular-nums`.
   - `bool`: status pill/badge.
   - `date` / `datetime`: formatted short date / date+time.
   - `enum`: deterministic tinted badge per value.
   - `belongs-to`: target's display field, linked when `Show.Href` is set.
   - `has-many as chips`: horizontal linked chip row.
   - `optional` empty values: omitted on cards, muted `—` in tables and details.
   - Named gap: `money` is deferred (philote has no `money` scalar type yet; `booking` uses `price: int`).

4. **Fixed themes (4)**:
   - Hand-tuned, contrast-verified light + dark CSS token sets (colors, radius, fonts): `plain` (default), `warm`, `editorial`, and `slate`.

5. **Spec mapping (21/21 mapped)**:
   - `board`: `ProjectCard` -> `card`; `ProjectView`, `TaskView` -> `detail`.
   - `library`: `AuthorCard` -> `card`; `AuthorView` -> `detail`; `BookShelf` -> `table`.
   - `neighbr`: `Landing` -> `hero`; `HorsePicker` -> `picker`; `JoinHerd` -> `callout`; `SwipeDeck` -> `deck`; `HorseCard`, `MatchCard` -> `card`; `HorseView` -> `detail`.
   - `booking`: `Hero` -> `hero`; `ClientPicker`, `SlotPicker` -> `picker`; `JoinClients` -> `callout`; `ServiceCard`, `BookingCard` -> `card`; `ClientView`, `ServiceView` -> `detail`.

## Consequences

- Frontend builds require zero model calls and produce zero frontend entries in `philote.lock` (only backend steps remain in `philote.lock`).
- `syntax`, `check`, `ir`, `lower`, `target/nextjs`, `docs/language.md`, and all golden specs (`board`, `library`, `neighbr`, `booking`) will be updated to use `: <kind>` and `theme:`.
- A future milestone can add a model-written component kind (e.g. `: custom`) without reworking pages, nav, or themes.

