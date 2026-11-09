+++
id = "I-003"
title = "AI writes the source, people review it, the checker is the AI's feedback"
status = "exploring"
summary = "A model turns a prose brief into source and iterates on checker errors until clean. A person reviews the short source diff. The compiler does the rest."
created = 2026-11-09
related = ["P-008", "Q-008", "E-002"]
+++

There are two LLM steps with a person between them:

1. Prose → source (AI author, iterating against `check`).
2. Source → code (pinned compile, P-003).

Error messages with hints are written for people, but they also suit a model fixing its own output. If the loop converges in a few rounds and the result is easy to review, that's a strong answer to Q-008.
