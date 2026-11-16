+++
id = "D-021"
title = "The bet is a deterministic compiler from spec to code"
status = "accepted"
summary = "Philote's value over a SPEC.md and an agent is determinism: the same spec builds the same app, and an edit moves only what it touches. Retargeting and upgrades are side effects."
created = 2026-11-12
answers = ["Q-008"]
related = ["P-003", "P-005", "P-008", "I-003", "E-001", "E-002"]
+++

## Decision

Philote is an AI compiler from spec to code, and its main value is less wiggle in the output. Building the same spec twice gives the same app. Editing a spec changes only the code that edit touches. An agent working from a SPEC.md can't promise either: every run can rewrite anything.

The other candidates in Q-008 rank under it:

1. **Short diff review** and **AI writes the source** (P-008) follow from it. A source diff is only worth reviewing if the compiler turns it into the same code every time.
2. **Retargeting** and **model upgrades** are side effects. They're welcome, but no design trade-off is made for them over determinism.

When two designs are otherwise equal, pick the one where a spec edit leaves more of the pinned output byte for byte the same.

## Why

The user's answer to Q-008: "when I generate a website I don't need it to be completely different each time". That's the gap between prompting an agent and compiling, and it's the one thing a model alone can't give.
