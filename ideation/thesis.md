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

## Where things live

Everything else in this folder is one item per file: [principles](principles/), [questions](questions/), [decisions](decisions/), [ideas](ideas/), [experiments](experiments/) and [prior art](prior-art/). [INDEX.md](INDEX.md) lists them all. The original ideation doc is kept as-is in [archive/](archive/).
