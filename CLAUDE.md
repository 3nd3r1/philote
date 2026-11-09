# philote

Ideation phase. Read `ideation/thesis.md` and `ideation/INDEX.md` first. The README explains the item format.

- Record ideas as items in `ideation/`, one per file. Create files with `make new KIND=... TITLE="..."`, never by guessing the next id.
- Any decision or principle you add or change is `proposed`. Never set `accepted`, and never mark a question `answered`, unless the user says so in this conversation.
- When the user accepts a decision, set it to `accepted` and mark each question in its `answers` as `status = "answered"` with `answered_by`.
- After editing items, run `make index` then `make check`. Never edit `ideation/INDEX.md` by hand.
- Don't delete items. Set them to rejected, retired, superseded or dropped, and say why in the body.
- Keep summaries plain and short (the cap is 200 chars). Reasoning goes in the body.
- `ideation/archive/` holds original documents. Don't edit them.
- No product code until the user asks. `prototypes/` holds throwaway experiments, each tied to an experiment item.
- Commits: conventional prefix (`feat:`, `fix:`, `chore:`, `docs:`, `build:`, optional scope like `docs(ideation):`), a short lowercase imperative subject, no body, no co-authors or trailers. One logical change per commit.
