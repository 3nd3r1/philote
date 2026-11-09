+++
id = "P-008"
title = "People review the source, not the code"
status = "proposed"
summary = "AI may write the source, but a person reviews every source change. The source must stay short and readable enough that this review is quick."
created = 2026-11-09
related = ["I-003", "I-004", "Q-008", "Q-016"]
+++

AI may write the source, but a person reviews every source change. The compiler guarantees that the code matches the source, so reviewing the source is enough.

This puts a constraint on the language: a source diff has to be readable in a minute. That pushes toward near-English syntax (Q-016) and tools that render the source back in plain terms (I-004).
