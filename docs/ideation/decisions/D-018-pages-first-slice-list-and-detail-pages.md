+++
id = "D-018"
title = "Pages first slice: list and detail pages, template components, relation reads"
status = "accepted"
summary = "A frontend declares list and detail pages over the backend's resources, and template components that show fields and relations. Resources gain read routes for their has-many relations. No model yet."
created = 2026-11-11
related = ["D-017", "D-016", "D-006", "D-004", "D-001", "Q-017", "Q-018"]
+++

## Decision

Pages and components go inside the frontend part, next to `uses`:

```
frontend web: nextjs {
  uses: api

  page "/horses" {
    data: horses
    layout: list of HorseCard
    empty: "No horses yet."
  }

  page "/horses/[id]" {
    data: horses by id
    layout: HorseView
    not_found: "That horse doesn't exist."
  }

  component HorseCard(horse: Horse) {
    shows: horse.name, horse.breed, horse.age
    links_to: "/horses/[horse.id]"
  }

  component MatchCard(match: Match) {
    shows: match.first, match.second, match.matched_at
  }
}

entity Horse {
  display: name
  ...
}
```

**Pages.** `data` names a resource of the used backend: `horses` lists it, `horses by id` loads one record by the route's `[id]`. A list page has `layout: list of <Component>` and must say `empty` (D-006). A detail page has `layout: <Component>`, needs `[id]` in its route, and must say `not_found`. The component's parameter must be the resource's entity.

**Components.** One parameter, an entity. `shows` lists what it displays, in order:

- `horse.name`: a field. Dates and datetimes are formatted, booleans print yes or no. Markdown can't be shown yet: that's a check error.
- `match.first`: a belongs-to relation, shown as the target's display field.
- `author.books as chips`: a has-many relation, shown as the display field of each record. A has-many relation without `as chips` is an error.

`links_to` makes the whole component a link to a page. Its route's params are fields of the parameter, like `[horse.id]`, and it must match a page.

**Display fields (D-017).** `display: <field>` on an entity names a string, slug, email or enum field. Showing a relation to an entity without one is an error with the hint `add display: name to Horse`. `display` becomes a reserved word inside entities.

**Reads (D-017).** `shows` is the component's read set. The generated props type has only what it shows, so later model-written bodies can't read more. Each page loads exactly those reads through the typed client:

- A belongs-to read calls the target resource's `get`.
- A has-many read calls a new relation route on the parameter's resource, like `horses.swipes(id)`.

Both need a resource in the used backend that serves that entity with `get`. If there is none, or there are several, that's a check error.

**Relation routes (backend).** Every resource with `get` also serves `GET /<resource>/[id]/<relation>` for each has-many relation of its entity, back sides included. It returns the related records, or `not_found` when the parent doesn't exist. `tree` and `expand` show these routes, and the client gets them as `horses.swipes(id)`.

**Generated files**, where the user's template keeps things:

- `src/app/<route>/page.tsx` for each page. A detail page also gets a `not-found.tsx` with its message.
- `src/components/common/<name>.tsx` for each component: a card that prints what `shows` lists. If it shows the entity's display field, that's the card's title.
- `src/lib/reads.ts`: each component's props type and the function that loads them.
- `src/lib/nav.ts`: the list pages, so the navbar links to them.
- `.prettierignore`, listing every generated file. The base adds shadcn's `card` and `badge` and a small `format.ts` for dates and booleans.

With Cache Components on, a page loads inside `<Suspense>` after `connection()`. So a missing record renders the `not_found` message with `noindex`, but the status is 200 (a "soft 404"). A real 404 status would need a Proxy check before streaming. That's left for later.

## Out of scope

Forms and actions (create, update, delete, swipe from the UI), model-written components, filters and ordering on lists, `as markdown`, the `theme` setting. Loads are one request per read per record, so a list of 50 matches makes 101 requests. Batching can come later without changing the source. A page also fails when a record it reads points at one that was deleted, since deletes don't check references yet (Q-018).

## Why

This is the smallest slice where the four D-017 answers all show up in generated code. Pages and components are template nodes, the look is the base's tokens, display fields are declared, and reads come only from the spec.

Pages read through the API rather than the data store, so the backend has to serve the reads pages need. Serving every has-many relation keeps the backend independent of any frontend. The other way, adding routes only when a frontend reads them, would make a component edit change the backend.
