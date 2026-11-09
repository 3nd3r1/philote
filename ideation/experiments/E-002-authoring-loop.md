+++
id = "E-002"
title = "Can a model write a clean, correct source from a prose brief?"
status = "planned"
summary = "Give a model a prose brief and the checker. Count rounds to a clean check, then judge whether the result is correct and quick to review."
created = 2026-11-09
tests = ["I-003", "P-008", "Q-008"]
+++

## Steps

1. Write three short prose briefs: the blog, the blog with comments and authors, and something unlike a blog.
2. Give a model the language reference and the checker. Let it write the source and iterate on errors.
3. Record rounds to a clean check and the errors it hit.
4. Review each result by hand: time the review, and list everything that's wrong despite passing the check.

## Measures

Rounds to clean, review time, wrong-but-valid mistakes per brief.

## Kill criteria

- The model regularly produces sources that pass the check but say the wrong thing, and the errors are hard to spot in review. That would mean strictness isn't buying reviewability.
