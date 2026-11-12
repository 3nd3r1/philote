+++
id = "D-020"
title = "Forms and actions first slice"
status = "proposed"
summary = "A page can be a form over a write endpoint. Components list actions: open a page, delete, or call an endpoint with bound inputs. Deletes refuse with in_use while referenced."
created = 2026-11-12
answers = ["Q-018", "Q-007"]
related = ["D-006", "D-017", "D-018", "D-019", "P-001", "P-005", "Q-004"]
+++

## Decision

**Form pages.** A page names a write endpoint with `form` instead of `data` and `layout`:

```
page "/horses/new" {
  form: horses.create
  then: "/horses/[id]"
}

page "/horses/[id]/edit" {
  form: horses.update by id
  then: "/horses/[id]"
  not_found: "That horse doesn't exist."
}

page "/swipe" {
  form: swipe
  then: "/matches"
}
```

- `form` names a resource's `create` or `update`, or a custom endpoint that takes an input. `update` needs `by id`, `[id]` in the route, and `not_found`, like a page by id.
- The fields are the endpoint's input, in order. Each field's kind picks its input: text, a text area for strings over 160 characters, email, number with its bounds, checkbox, date, date and time, a select of the enum's values, and a select of records for a foreign key or record input, labelled by the target's display field. A record select needs the target's `display` and one resource that lists it.
- `then` is required: the page to go to after a save. Its `[id]` is the saved record's id, so the endpoint must return a record of the entity that page loads.
- The form is a template. It posts to a server action, which calls the typed client. `invalid_input` issues show under their fields. Other declared errors show above the submit button. The form works without JavaScript.
- Form pages without `[id]` join the nav, after the list pages.

**Actions on components.** A component lists what can be done to its record:

```
component HorseView(horse: Horse) {
  shows: horse.name, horse.bio
  actions: open "/horses/[horse.id]/edit", delete then "/horses"
}

component BookCard(book: Book) {
  shows: book.title
  actions: mark_out_of_print(book: book, in_print: false) as "Out of print"
}
```

- `open "<route>"` is a link to a page, filled like `links_to`.
- `delete then "<route>"` deletes the record through the one resource that deletes its entity.
- `<endpoint>(<input>: <value>, ...)` calls a custom endpoint. Every input must be bound to the record, one of its fields or belongs-to relations, or a literal. Inputs can't come from who is signed in yet (Q-004), so an endpoint like `swipe`, which needs the acting horse, is a form for now. `then` is optional; without it the page reloads.
- `as "<label>"` sets a button's text. The defaults are the target page's title, "Delete", and the endpoint's name.
- Actions are rendered by the template beside the model-written body, not inside it. Adding or changing an action doesn't change the component's prompt, so its pinned body stays as it is (P-005). A component with `links_to` puts its actions under the link.

**Deletes refuse while something points at the record (Q-018).** A resource's `delete` gains the declared error `in_use` (409) when its entity has a has-many side. The backend refuses while any record points at the one being deleted. Deleting a record that's already gone counts as done in the UI.

**Error copy.** Built-in errors have fixed text from the template. A custom error shows as its name in words, like "Same horse." Spec copy for errors can come later without breaking specs.

**Verification (Q-007).** Generated apps are checked with types and smoke tests, not generated end-to-end tests. Golden files and `tsc` cover every case. The smoke test also posts the forms the way a browser without JavaScript would.

## Why

A form is a page whose data goes the other way, so it reuses everything pages already check: routes, `by id`, `not_found`, the nav. Fields come from the endpoint's input, so a form can't send what the backend doesn't accept. Forms stay templates because they carry the data contract. A wrong field name or type is a broken app, not a design choice.

Actions sit on components because that's where the record is. The model's body never sees them, which keeps every action change out of the lock.

Deletes are templates, so no model looks at them and the rule has to be written down once. Refusing is the strict default (P-001). A per-relation cascade can come later without changing what existing specs mean.
