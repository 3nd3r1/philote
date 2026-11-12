+++
id = "Q-018"
title = "What happens to records that point at a deleted record?"
status = "answered"
answered_by = "D-020"
summary = "Deleting a Horse leaves its Swipes and Matches pointing at an id that no longer exists, and pages that show them fail. Refuse the delete, delete the dependents, or something declared per relation?"
created = 2026-11-11
related = ["D-014", "D-018"]
+++

Today `DELETE /horses/[id]` always succeeds. A Match whose `first_id` named that horse now points at nothing. The backend still lists it, and a page showing `match.first` fails when it loads the missing horse.

## Options

- **Refuse.** A delete fails with a new declared error, `in_use`, while anything points at the record. Nothing is lost by accident, and the client gets the error in its union.
- **Cascade.** Deleting a horse deletes its swipes and matches too. Convenient, but one call can remove a lot.
- **Declared per relation.** `Swipe belongs to target: Horse (on delete: cascade)`, with refuse as the default. The most precise, and one more thing to write.

Refuse as the default fits P-001: nothing happens that the spec didn't say. A per-relation cascade could come later without breaking specs that rely on the default.
