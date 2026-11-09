+++
id = "D-004"
title = "Dependencies come from what the checker resolves"
status = "proposed"
summary = "A node depends on exactly the nodes whose names and fields it resolves during checking, so dependencies are tight without being declared."
created = 2026-11-09
related = ["P-005", "Q-012"]
+++

## Decision

While type-checking, the checker records every entity, field, component and page each node resolves. That set is the node's dependency list.

## Why

It's tight without extra syntax: PostView reads `post.tags` as chips, so it depends on Tag; PostCard doesn't, so adding a field to Tag leaves PostCard cached. The E-001 prototype checks these cases.
