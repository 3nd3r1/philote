+++
id = "D-011"
title = "Focus on spec to code first"
status = "accepted"
summary = "Build the CLI and tooling that turn a spec into code. Generating specs with AI comes later. Pages and components are not the focus; backends are."
created = 2026-11-09
related = ["E-002", "I-009"]
+++

## Decision

Next work is the `philote` CLI going from spec to code. AI-written specs (E-002) wait. The blog-and-components shape was too narrow; the test case should include a backend.

## Why

The product is the compiler. E-002 showed a model can write valid specs for what the language supports, so that's not the risk right now.
