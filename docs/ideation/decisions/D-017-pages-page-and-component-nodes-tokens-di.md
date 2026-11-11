+++
id = "D-017"
title = "Pages: page and component nodes, tokens, display fields, declared reads"
status = "accepted"
summary = "Pages and components are the frontend's nodes. The look comes from theme tokens. An entity shown by reference declares its display field. A component's reads are part of its interface."
created = 2026-11-11
answers = ["Q-003", "Q-005", "Q-011", "Q-012"]
related = ["D-001", "D-002", "D-004", "D-015", "D-016", "P-007", "I-006", "Q-017"]
+++

## Decision

1. **Nodes are pages and components, nothing smaller (Q-005).** A page is a template node: it loads data through the typed client (D-016) and handles not found and empty. A component has props the compiler fixes, like `HorseCard(horse: Horse)`. By default it's a template built from its `shows` line. Only a component marked custom gets a model-written body, the way steps do (D-015).
2. **The look is theme tokens, for now (Q-003).** The frontend names a shadcn preset, like `theme: "b4pl265hL6"`, and templates compose the base's shadcn components. Visuals that need more go through custom components or native blocks (I-006). Q-017 keeps the bigger question open.
3. **An entity shown by reference declares its display field (Q-011).** `entity Horse { display: name ... }`. Showing a Horse as a chip, link or picker option without it is a check error, with the hint `add display: name to Horse`. There is no implicit rule.
4. **A component's reads are part of its interface (Q-012).** A template component's `shows` line is its read set. A custom component declares its reads the same way, and its props type only contains what it declared. The page computes its loads from those declarations, never from a model-written body, so a body can't change what its page loads.

## Why

Each answer keeps the compiler owning the interfaces (D-001) and the model inside small boxes.

Smaller nodes than components would add many interfaces for little saved per rebuild. A component with fixed props and a model-written body is the frontend's version of a step, which already works.

A design node would make every page depend on one large model output, so changing it would rebuild everything. Tokens keep the look deterministic and local. The user's template already treats a preset code as the one place the look comes from.

An implicit display rule is a hidden rule (P-007). Choosing the field at each use repeats it in every chip, link and picker. One declaration on the entity covers them all.

Reads declared in the spec are what the checker already resolves (D-004). Computing loads from them keeps page loaders templates, and keeps model output from leaking into data loading.

## Consequences

Loading a relation through the API needs relation reads on the backend, like `GET /horses/[id]/matches` for a has-many relation. Today a page could only list everything and filter. That gets its own decision with the first pages slice.
