+++
id = "P-004"
title = "Verification comes from the source, not the LLM"
status = "accepted"
summary = "Types, routes, props and data shapes are checked mechanically against the source. A node that fails is regenerated, or the build fails."
created = 2026-11-09
related = ["Q-007"]
+++

Types, routes, props and data shapes are checked mechanically against the source. A node that fails gets regenerated, or the build fails, with the errors shown against the source line, not the generated code.
