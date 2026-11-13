+++
id = "Q-020"
title = "How does a spec mark a field the app sets, not the user?"
status = "open"
summary = "A Task's finished_at is set by a step, but the create form and POST /tasks still take it. How does a field say that only steps write it?"
created = 2026-11-13
related = ["D-020", "D-022", "P-001", "P-007"]
+++

Found by building a task board from scratch. `finished_at: datetime optional` is set by the `finish_task` step, but `tasks.create` and `tasks.update` take it as input, so the New Task form shows a "Finished at" picker. Slugs are already left out of input, because the template makes them. Nothing else can be.

## Options

- **A modifier.** `finished_at: datetime optional set_by_steps`, or a shorter word. The field leaves every resource's input and form. Only steps write it. The checker can require that some step writes the entity, so the field isn't dead.
- **Per-operation input.** `resource tasks: Task { create: title, status, project }`. Exact, but every resource repeats its field list, and an added field is left out until someone adds it here.
- **Leave it.** A form can't hide a field, so the user sees fields they shouldn't fill.

A modifier fits P-001 best: it's one word on the field, and both the input checks and the form follow from it. I'd pick a modifier, and decide its name with the user.
