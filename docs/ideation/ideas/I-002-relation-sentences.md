+++
id = "I-002"
title = "Relation sentences with made-up verbs"
status = "raw"
summary = "Write relations as short sentences (`Post has many Comment`, `User follows User`). Made-up verbs become named edges the compiler can store and check."
created = 2026-11-09
related = ["P-007", "Q-015", "I-005"]
+++

Sketch, syntax not settled:

```
Post has many Comment
Comment belongs to User as author
User follows User                   // made-up verb: a named many-to-many edge
User likes Post
User can delete Comment if author   // permission, checked against the relation
note "follows are public, like Twitter"
```

Built-in verbs have fixed meaning. A made-up verb is a label on an edge: the compiler doesn't need to know what "likes" means to store it, query it and check every use of it.
