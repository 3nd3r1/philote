+++
id = "Q-017"
title = "How much design freedom does the model get?"
status = "answered"
answered_by = "D-019"
summary = "Tokens keep the look deterministic but narrow. The point of using a model is that it can build anything. Where can it design freely while builds stay pinned?"
created = 2026-11-11
related = ["D-017", "D-015", "I-006", "P-007"]
+++

D-017 starts with theme tokens. The worry: if every page is a template over a fixed component library, philote is just a programming language, and the model's main strength, building anything, goes unused.

Determinism comes from pinning, not from templates. A step body is model-written and still deterministic, because the lock holds it. So the real costs of more design freedom are locality (a change rebuilds more) and reviewability (what's in the spec says less about what you get), not determinism.

## Options

- **Custom components only.** The model designs freely inside one component's box, with props the compiler fixes. This already works under D-017.
- **A pinned design node.** The model writes a design once (layout, component styling) and it's pinned. Every page depends on it.
- **Design per page.** A page can be marked custom. The model lays it out from the components it's allowed to use, and the page's data stays compiler-owned.
- **Design intent as a capped field.** Like a step's `intent`, a short style note that reaches the model. This conflicts with P-007 unless it's treated as a deliberate exception.
