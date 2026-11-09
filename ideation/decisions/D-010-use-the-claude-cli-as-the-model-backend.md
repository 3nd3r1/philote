+++
id = "D-010"
title = "Use the claude CLI as the model backend for experiments"
status = "accepted"
summary = "Experiments call `claude -p` behind one small function. Switch to the API once the idea holds up, or if the CLI is too slow or unpredictable."
created = 2026-11-09
answers = ["Q-014"]
+++

## Decision

Experiments call the `claude` CLI, wrapped in one function so swapping to the API is a one-file change. The CLI version goes into the cache key.

## Why

No setup and no bill. The CLI adds hidden instructions we don't control, but pinned outputs don't change once cached, so that matters less for experiments. A real compiler will need the API.
