+++
id = "I-009"
title = "Zoom levels: a tree of nodes, lowered one level at a time"
status = "exploring"
summary = "The app is a tree: system, then frontend/backend, then resources, then endpoints, then steps. Stop at any level; what's below is generated, pinned, and viewable as spec."
created = 2026-11-09
related = ["D-001", "D-002", "P-007"]
+++

The app is a tree of nodes, each with a path (`shop/backend/checkout`):

    shop                          system: frontend, backend, data
      backend api                 resources and endpoints
        resource orders: Order    CRUD, expanded by a template
        endpoint checkout         POST /checkout, input, output, errors
          steps                   reserve stock, charge, create Order, send receipt

You can stop at any level. Whatever is below is filled in:

- **by a template** when it's mechanical (`resource orders: Order` expands to five endpoints), or
- **by the model** when it isn't. The model writes the next level down *as spec*, not code, and that's pinned too.

Either way, `philote expand <path>` shows the filled-in level as spec, and you can copy it into your source to edit it. Zooming in moves a node from "described" to "specified". So zoom is the dial between AI and determinism: the more you specify, the fewer model calls.

Backend endpoints are the most deterministic part. Route, method, input and output types, validation, status codes, error shapes, OpenAPI and a client are all templates. Only custom steps need the model.

## File layout

Every node can live inline in its parent or in its own file at its path, like Rust modules. `philote tree` shows the whole tree at any depth, so many small files stay navigable.

Prior art: MLIR's progressive lowering through dialects (from memory).
