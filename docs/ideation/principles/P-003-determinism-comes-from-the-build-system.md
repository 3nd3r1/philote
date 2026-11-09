+++
id = "P-003"
title = "Determinism comes from the build system, not the model"
status = "accepted"
summary = "Every node's output is pinned by a hash of its inputs. Identical inputs never call the model."
created = 2026-11-09
related = ["I-007", "D-003"]
+++

Every node's output is pinned by a hash of its inputs. Identical inputs never call the model.

This is pinned replay, not true determinism: a cached output is one sample, not the only possible one.
