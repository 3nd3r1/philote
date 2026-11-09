+++
id = "P-007"
title = "Every word has a checkable meaning"
status = "proposed"
summary = "A word in the source is a built-in term, a name you made up, or a note the compiler ignores. No words that only the model interprets."
created = 2026-11-09
related = ["P-001", "I-002", "Q-015"]
+++

Words can go into the source as long as each one is something the compiler can check. Three kinds:

1. **Words the compiler knows**, like `has many`, `belongs to`, `can … if`. Each compiles to something concrete (a foreign key, a join table, a permission check) and gets checked.
2. **Words you make up**, like `follows`, `interacts with`, `mentions`. The compiler doesn't need to know what "follows" means: `User follows User` is a named edge between two entities that it can store, query and check. The word is just a label.
3. **Notes**: free text the compiler ignores. Shown during review, never sent to codegen.

There is no fourth kind: words only the model interprets. "User interacts with Post" alone doesn't say what is stored, who can do it, or where it shows up, so the checker should ask for that structure instead of letting the model guess.

Open edge: P-001 already allows "a short intent hint" that does reach the model. Either that hint becomes a note (never reaches codegen), or it's a deliberate, capped exception to this principle.
