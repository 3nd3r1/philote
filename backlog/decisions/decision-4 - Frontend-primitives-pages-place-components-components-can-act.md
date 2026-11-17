---
id: decision-4
title: 'Frontend primitives: pages place components, components can act'
date: '2026-11-17T10:36:00Z'
status: proposed
---
## Context

The frontend has three fixed page kinds (list, record, form), a fixed nav, and a template that owns every page's structure. The model only fills holes, so every app is the same admin panel, and a swipe deck or a landing page can't be said at all (decision-3). Determinism doesn't need any of that: it comes from the lock. The template only has to own what must be exact: data loading, the typed client, routes, actions and forms.

## Decision (proposed)

Five primitives replace the page kinds.

**1. A page loads data and places components.** No kinds, no template heading or grid. The page renders its components in order; each component is a full-width section that owns its own layout.

```
page "/horses/[id]" {
  load: horse from horses by id
  shows: HorseView(horse)
  not_found: "That horse doesn't exist."
}
```

`load` names what the page fetches: `horses` (a resource's list), `horses by id` (one, from the route), or a custom read endpoint like `candidates(horse: viewer)`. `empty` and `not_found` stay on the page. `/` becomes a normal route.

**2. A component takes records or lists, and can be composed.** Parameters can be `list of Horse`. `uses: HorseCard` lets a component render another one. `shows: list of HorseCard(horses)` on a page is shorthand for a plain grid of cards.

**3. A component can act.** `calls: swipe(horse: viewer, target: deck, liked)` hands the component a typed action for that endpoint. Inputs bound in the spec are filled in by the template; inputs left open (`liked`) come from the UI. A component with `calls` is interactive (a client component). It can't fetch anything else: the template passes it the action, and the check rejects `fetch`.

**4. The spec owns the words.** `says: "Find your herd.", "Swipe on horses near you."` gives a component its copy, and the check fails if any of it is missing from the body. Text a user reads comes from data, `says`, or an `as` label; the model adds icons and layout, never words.

**5. The app shell is in the spec.** `nav: bottom, "/discover" as "Discover", "/matches" as "Matches"` (or `top`, `side`). The template passes the items and the active one; the model draws the nav from the app's look, pinned like a component.

Plus one concept the swipe needs: **`viewer: Horse`**. The app lets you pick which horse you are (a cookie, no login). Loads and calls can use `viewer`; a page that needs it sends you to the picker first. Auth (Q-004) later replaces the picker without changing specs.

### neighbr in the primitives

```
// api.philote: today's resources, swipe and its steps, plus two reads
endpoint candidates {
  route: GET "/horses/[horse]/candidates"
  output: list of Horse
  steps: pick_candidates
}

step pick_candidates(horse: Horse) -> horses: list of Horse {
  reads: Horse, Swipe
  intent: "every other horse this one hasn't swiped on yet"
}

endpoint matches_of {
  route: GET "/horses/[horse]/matches"
  output: list of Match
  steps: find_matches
}

step find_matches(horse: Horse) -> matches: list of Match {
  reads: Match
  intent: "every match this horse is in, newest first"
}
```

```
// web.philote
uses: api
look: "a warm, playful dating app: coral and cream, soft rounded cards, friendly rounded type"
viewer: Horse
nav: bottom, "/discover" as "Discover", "/matches" as "Matches", "/me" as "Me"

page "/" {
  shows: Landing
}

page "/pick" {
  load: horses
  shows: HorsePicker(horses)
  empty: "No horses yet."
}

page "/discover" {
  load: deck from candidates(horse: viewer)
  shows: SwipeDeck(viewer, deck)
  empty: "You've seen every horse. Check back later."
}

page "/matches" {
  load: matches from matches_of(horse: viewer)
  shows: list of MatchCard(matches)
  empty: "No matches yet. Keep swiping."
}

page "/me" {
  shows: HorseView(viewer)
}

page "/horses/[id]" {
  load: horse from horses by id
  shows: HorseView(horse)
  not_found: "That horse doesn't exist."
}

page "/horses/new" {
  form: horses.create
  then: "/pick"
}

component Landing {
  says: "Find your herd.", "Swipe on horses near you. Match when it's mutual."
  links_to: "/pick" as "Start swiping"
  look: "a full-screen hero: the headline huge, the line under it, one big button"
}

component HorsePicker(horses: list of Horse) {
  shows: horses.name, horses.breed
  picks: viewer then "/discover"
  says: "Who are you?"
  look: "a grid of round avatars with the initial and the name; tap one"
}

component SwipeDeck(viewer: Horse, deck: list of Horse) {
  shows: deck.name, deck.age, deck.breed, deck.bio
  calls: swipe(horse: viewer, target: deck, liked)
  look: "tinder: one big card at a time, drag right to like, left to pass, round ✕ and ♥ buttons under it, the next card peeking behind"
}

component MatchCard(match: Match) { ...as today }
component HorseView(horse: Horse) { ...as today }
```

### Today's pages in the primitives

- List: `data: horses` + `layout: list of HorseCard` becomes `load: horses` + `shows: list of HorseCard(horses)`.
- Record: `data: horses by id` + `layout: HorseView` becomes `load: horse from horses by id` + `shows: HorseView(horse)`.
- Form: unchanged. Forms stay templates (they carry the data contract, D-020), styled by the theme, and a form page can also `shows` components above the form.
- decision-2's model-written page layout goes away: pages are a stack of components, and components own their layout.

## Consequences

- New language: `load`, `shows`, `list of` parameters, `uses`, `calls`, `says`, `picks`, `nav`, `viewer`; GET endpoints with route inputs and `list of` outputs. docs/language.md and golden cases for each, and library and board are rewritten.
- The model writes more of the app: every section, the nav, and interactive components. Each is still pinned and only rewritten when its own prompt changes; checks are tsc, ESLint, `says` present, no `fetch`.
- Interactive components can't be checked by smoke tests without a browser. A Playwright smoke run would need a new dev dependency, the user's call.
- Built in slices: (1) load/shows/list of/uses and `/` + says, (2) nav, (3) viewer and picks, (4) calls and GET list endpoints, then neighbr side by side with prompting Claude directly.
