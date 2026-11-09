+++
id = "Q-006"
title = "Regeneration on change: fresh sample or minimal edit?"
status = "answered"
answered_by = "D-015"
summary = "When a node's source changes, does the model start fresh, or edit the previous output?"
created = 2026-11-09
related = ["I-007"]
+++

The original doc leans toward minimal edit: the previous output is passed in as a reference, so the model changes as little as possible. The cost is that the cache key then depends on the previous output too, which makes builds path-dependent: the same source can produce different code depending on its history.
