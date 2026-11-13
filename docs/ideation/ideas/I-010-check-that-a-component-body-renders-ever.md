+++
id = "I-010"
title = "Check that a component body renders every field it shows"
status = "raw"
summary = "tsc proves a body uses fields that exist, not that it uses every field in shows. A check that each shown field appears in the body would catch a model that drops one, and retry like a type error."
created = 2026-11-13
related = ["D-019", "P-004", "E-003"]
+++

In E-003, adding `horse.color` to HorseView's `shows` made the model reuse the slot of a duplicate breed line for the color. That was fine, because breed still showed as a badge. But nothing checked it. A body that drops a shown field typechecks and gets pinned.

The check could be mechanical: each path in `shows` must appear in the body, like `horse.color`. A miss goes back to the model with the missing fields, like a tsc error (D-015).
