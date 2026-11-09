+++
id = "Q-011"
title = "How does a list of records pick the field to show?"
status = "open"
summary = "`post.tags as chips` doesn't say which Tag field to show. Implicit rule, explicit field, or a display field declared on the entity?"
created = 2026-11-09
related = ["P-001", "P-007"]
+++

## Options

- **Implicit rule** (what the prototype does): use the entity's only string field; error if there are several. Works for the blog, but it's a hidden rule.
- **Explicit at the use site**: `post.tags as chips(name)`. Strict, slightly noisier.
- **Declared on the entity**: `entity Tag { display: name ... }`. One place, reused everywhere.
