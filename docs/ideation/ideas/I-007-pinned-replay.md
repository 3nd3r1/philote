+++
id = "I-007"
title = "Pinned replay: lockfile, content-addressed store, frozen builds"
status = "adopted"
summary = "Each node's output is keyed by a hash of its inputs and kept in a store, like /nix/store. Normal builds regenerate dirty nodes only; frozen builds never call a model."
created = 2026-11-09
related = ["P-003", "D-003", "Q-006", "A-007"]
+++

The cache key covers the node's own source, the interface hashes of its dependencies (not their generated code), the backend target and version, the model id and parameters, and the prompt template and compiler versions. The lockfile maps each key to an output hash; outputs live in a content-addressed store.

- **Normal build:** regenerate only nodes whose key changed.
- **Frozen build** (`--frozen`, for CI): never call a model; fail if anything is dirty.
- **Upgrade** (like `nix flake update`): bump the model or backend on purpose, rebuild everything, review the diff.

Pipeline: parse → check → hash nodes → compare with lockfile → clean nodes come from the store, dirty nodes go to the model → verify against the source (retry a few times) → pin. A node that keeps failing fails the build, with errors shown against the source line.

To keep a changed node from drifting: pass the previous output as a reference (Q-006), and reject any output that breaks the node's interface (D-001).
