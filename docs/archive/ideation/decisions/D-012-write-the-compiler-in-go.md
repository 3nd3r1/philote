+++
id = "D-012"
title = "Write the compiler in Go"
status = "accepted"
summary = "The real compiler is Go, standard library first. Templates use text/template and embed."
created = 2026-11-09
answers = ["Q-009"]
supersedes = ["D-005"]
+++

## Decision

Write philote in Go.

## Why

Installed, fast builds, and the prototype is already Go. `text/template` and `embed` cover the template-first targets with no dependencies. A plain language leaves less room for clever, hard-to-read code.
