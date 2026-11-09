+++
id = "I-005"
title = "Flag relations that are stored but never used"
status = "raw"
summary = "If `User likes Post` exists but no action changes it and no view shows it, that's an error: the source says something exists without saying what it's for."
created = 2026-11-09
related = ["I-002", "P-007"]
+++

```
error: 'User likes Post' is stored, but nothing shows or changes it
  = hint: add an action (e.g. `action like(post: Post)`) or show it (`post.likes as count`)
```
