# Thesis

Code becomes a compile target. You write a strict, typed description of the app (entities, relations, pages, components, actions). A compiler checks it, then uses an LLM as its backend to emit framework code. The generated code is machine code: never edited, pinned in a lockfile, and rebuilt only where the source changed.

Why strict objects instead of prose specs: prose specs drift and fill up with AI slop. A typechecked source can still be wrong, but it can't be vague. Ambiguity fails at compile time instead of turning into a guess buried in generated code.

The bet: the right level for AI codegen is a high-level language, not a chat window and not a Markdown spec.

## The workflow it enables

The source is an ontology of the product: concepts, the relations between them, and the views onto them. That makes a two-stage loop possible:

1. A person describes the product in prose. An AI writes or edits the source.
2. The AI runs the checker and fixes errors until the source is clean. Errors with hints are good feedback for a model.
3. The person reviews the source diff: a few lines of near-English instead of hundreds of lines of code.
4. The compiler builds code from the reviewed source. Builds are pinned, so the same source always gives the same code, and an edit only rebuilds what it touches.

People review the ontology; the compiler guarantees the code matches it.

## Principles

1. **The source is strict.** Every object has a type and a fixed schema. Natural language only in named, length-capped fields.
2. **Generated code is machine code.** It's never edited by hand. Changes go into the source or a native block.
3. **Determinism comes from the build system.** Every node's output is pinned by a hash of its inputs. Identical inputs never call the model.
4. **Verification comes from the source.** Types, routes and data shapes are checked mechanically. A node that fails is regenerated, or the build fails.
5. **Regeneration stays local.** Changing one node rebuilds that node only, unless its interface changed.
6. **Frameworks are backends.** The same source should compile to a second framework without edits.
7. **Every word has a checkable meaning.** A word is a built-in term, a name you made up, or a note the compiler ignores.
8. **People review the source, not the code.** A source diff has to be readable in a minute.

Work and direction live in the board (`backlog board`). The design history up to 2026-11-16 is frozen in [archive/ideation](archive/ideation/INDEX.md).
