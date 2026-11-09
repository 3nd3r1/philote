+++
id = "D-009"
title = "Use a JSON file as the data source"
status = "accepted"
summary = "Experiments read data from a JSON file. Data sources aren't what we're testing."
created = 2026-11-09
answers = ["Q-013"]
+++

## Decision

Experiments read their data from a JSON file.

## Why

Data sources don't matter for testing the idea, and a JSON file works unchanged on any framework.
