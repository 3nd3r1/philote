+++
id = "D-023"
title = "Readonly fields are written only by steps"
status = "proposed"
summary = "A field marked readonly leaves every resource's input and form. Only steps write it. It must be optional, and some step must write its entity."
created = 2026-11-15
answers = ["Q-020"]
related = ["D-020", "D-022", "P-001", "P-004", "P-007"]
+++

## Decision

A field can be marked `readonly`, after the type like `optional` and `unique`:

```
entity Task {
  display: title
  title: string(max 120)
  status: Status
  finished_at: datetime optional readonly
}
```

- A readonly field is left out of the input of the resource's `create` and `update`, so it's not in the API body or in a form over them. It still shows up in every read.
- Only steps write it. A step that lists the entity in `writes` can set it, like any other field.
- It must be `optional`. A resource creates the record without it, so a required one could never be filled. Saying `readonly` without `optional` is a check error, with a hint to add `optional`.
- Some step must list the entity in `writes`. Otherwise nothing ever sets the field, and the checker says so at the field.
- A `slug from` field is already left out of input, so `readonly` on it is a check error.
- A custom endpoint's `input` is its own field list, so it isn't affected.

## Why

Q-020: the task board's New Task form had a "Finished at" picker for a field only `finish_task` sets. A modifier is one word on the field, and both the input type and the form follow from it (P-001). Who writes the field is stated in the spec (P-007), and the two checks catch a field that can't be filled (P-004).

`readonly` is the word people know from forms and APIs. It describes the field from the user's side, which is the side the spec is about.

The language has no default values, so requiring `optional` is the simple rule for now. If defaults come later, a readonly field with a default can drop it.
