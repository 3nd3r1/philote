+++
id = "D-024"
title = "A shown relation links to its one page"
status = "accepted"
summary = "If exactly one page loads an entity by id, every shown relation to that entity links there. The link comes in the props as an href, so the template decides it and the model only renders it."
created = 2026-11-15
answers = ["Q-021"]
related = ["D-017", "D-018", "D-019", "D-020", "Q-011", "P-004"]
+++

## Decision

A component that shows a relation links each related record to that entity's page, when the frontend has exactly one page with `data: <resource> by id` over it. Nothing new to write.

```
page "/tasks/[id]" { data: tasks by id  layout: TaskView }

component ProjectView(project: Project) {
  shows: project.name, project.tasks as chips   // each chip links to /tasks/[id]
}
```

- The related record in the props gets an `href: string` next to its `id` and display field. The read function fills it in, so the template decides where it points. The prompt says to render a record that has an `href` as a link to it.
- No page, or two pages over the entity: no `href`, and the record is plain text, like today.
- A component with `links_to` is already one link, and a link can't sit inside another. Its related records get no `href`.
- `philote tree` prints each linked relation and where it goes, so a reviewer sees links without reading the bodies.
- A relation that starts linking changes the component's props and prompt, so its body is written again, starting from the old one.

## Why

Q-021: on the task board, `/tasks/[id]` could only be reached by typing the URL. The checker already knows every page and what it loads, so the link follows from the spec without the author writing it, and the same spec always gives the same links.

Declaring `links_to` on each relation would be the same line in almost every spec. Two pages over one entity are rare, and `tree` shows when a relation stays unlinked.

Putting the href in the props keeps the URL out of the model's hands: it can't link to a page that doesn't exist. Whether it actually renders the link is the same question as whether it renders every shown field, which I-010 would check.

## Progress

Built on main, 2026-11-15. Five relations across the three golden cases started linking, and four component bodies were rewritten. Each new body renders every href. BookCard came back with only its author line changed.
