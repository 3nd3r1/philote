+++
id = "D-003"
title = "Cache keys hash the canonical form, not the source text"
status = "proposed"
summary = "Each node is printed in a canonical form before hashing, so whitespace, comments and formatting never trigger a rebuild."
created = 2026-11-09
related = ["P-003", "I-007"]
+++

## Decision

A node's cache key hashes its canonical printed form (from the parsed source), the interface hashes of its dependencies, and the build config (target, compiler, model and prompt versions).

## Why

Formatting changes shouldn't cost model calls. The canonical printer is also a free formatter. The E-001 prototype checks this: a whitespace and comment edit changes no keys.
