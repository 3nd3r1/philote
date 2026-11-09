+++
id = "Q-005"
title = "Node granularity"
status = "open"
summary = "Is a build node a page, a component, or something smaller?"
created = 2026-11-09
related = ["D-001"]
+++

The prototype uses one node per entity, page and component. Smaller nodes mean cheaper rebuilds but more interfaces for the compiler to own (D-001).
