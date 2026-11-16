+++
id = "D-013"
title = "First targets: Next.js and Hono, as typed templates"
status = "accepted"
summary = "The first targets are a Next.js frontend and a separate Hono backend. Each is a real base project plus templates filled from the spec. The model fills only marked slots."
created = 2026-11-09
related = ["D-001", "D-002", "P-006", "I-009"]
+++

## Decision

Support exactly two targets first: a Next.js frontend and a Hono backend (TypeScript), as separate nodes. The frontend gets a generated, typed client for the backend.

Each target is a Go type that declares the node kinds it supports. It owns a runnable base project with pinned versions, and templates filled from the checked tree. Anything a target doesn't support is a check error. The model only fills slots a template marks, and only for nodes marked custom.

## Why

A small, fixed set of targets lets templates do almost everything, which makes builds deterministic by default. The model becomes the exception, not the engine.
