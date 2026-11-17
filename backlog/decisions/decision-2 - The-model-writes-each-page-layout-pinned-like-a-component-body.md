---
id: decision-2
title: The model writes each page layout, pinned like a component body
date: '2026-11-17T09:55:00Z'
status: accepted
---
## Context

After decision-1 every app had its own colors and fonts but the same pages: templates owned each page's structure (a title, then a two-column grid of cards in a narrow column), and the model wrote only the inside of each card. The user looked at neighbr on 2026-11-17 and called it horrible. Card polish can't fix a page's composition.

## Decision

The user chose this on 2026-11-17, over richer templates alone and over adding images to the language first.

- Every list, record and form page has a slot for its layout, PageBody. The model writes it; it's checked and pinned in philote.lock as `page web/<route>`, like a component body.
- The layout gets the page's title and a content(className) function. It sets the width, header and spacing, and the class says how the cards are arranged. Loading, empty and not-found states, data and forms stay templates.
- A layout's prompt is built from the app's name and look line and the page alone. Editing a component never rewrites a page; changing the app's look rewrites the theme and every layout, not the components.

## Consequences

Each page costs one model call the first time. `/theme` can't be a route, since `web/theme` names the theme. The navbar is still a base file; if apps still look alike, it's the next candidate.

