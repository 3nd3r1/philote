+++
id = "D-001"
title = "The compiler owns every interface; the model only fills bodies"
status = "accepted"
summary = "File paths, export names, prop signatures and imports are fixed by the compiler from the source. The model writes what goes inside."
created = 2026-11-09
related = ["P-005", "Q-005", "E-001"]
+++

## Decision

For every generated node, the compiler decides the file path, the exported name, the exact signature and the available imports, all derived from the source. The model only writes the body. Verification checks the body against that contract.

## Why

Locality (P-005) then holds by construction instead of depending on the model keeping signatures stable. It is the direct answer to E-001's first kill criterion (interface churn breaking neighbouring nodes). It also makes verification cheap: a node is valid if it type-checks against the stub the compiler expected.

## Consequences

Each backend needs a small deterministic layer that knows how to express an interface in its framework. That is hand-written per target (the Wasp cost, A-005), but it's small.
