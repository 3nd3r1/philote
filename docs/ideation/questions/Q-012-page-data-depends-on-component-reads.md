+++
id = "Q-012"
title = "What does a page load when its component reads relations?"
status = "open"
summary = "PostView shows post.tags, so /blog/[slug] must load tags. A page's data shape depends on what its components read, not just their interface."
created = 2026-11-09
related = ["D-001", "D-004"]
+++

## Options

- **Load each record with its relations one level deep.** Simple; over-fetches.
- **Compute the shape from what components read.** The checker already knows every field path a component touches (D-004), so it can derive the exact query. But then a component's body changes its page's data, which breaks locality unless the data loader is its own node.
- **Make the read set part of the component's interface.** Honest about the dependency, and the loader becomes a template node.
