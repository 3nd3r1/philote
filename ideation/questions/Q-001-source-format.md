+++
id = "Q-001"
title = "Source format: custom DSL or an existing format?"
status = "open"
summary = "A custom DSL (best error messages, more work) or an existing format like TOML, YAML or a Nix expression (free parser)?"
created = 2026-11-09
related = ["Q-016", "D-007"]
+++

## Options

- **Custom DSL.** Best error messages, can read like English (Q-016), more work.
- **TOML / YAML.** Free parser, familiar, but relations and queries end up as strings that need a second parser anyway.
- **Nix expression.** Free evaluation and composition, but unfamiliar to most people and hard for quick review.

## Notes

The step-1 prototype (E-001) used a placeholder DSL. Its parser was small: the hard part is the checker, which is the same work for any format. That weakens the "free parser" argument.
