+++
id = "Q-021"
title = "Does a related record link to its page?"
status = "open"
summary = "A project's tasks show as chips, and a task shows its project, but neither links to the record's page, so /tasks/[id] can only be reached by URL. Link by default, declare it, or never?"
created = 2026-11-13
related = ["D-017", "D-018", "Q-011"]
+++

Found by building a task board from scratch. ProjectView shows `project.tasks as chips`, and TaskView shows `task.project`. The app has pages at `/tasks/[id]` and `/projects/[id]`, but the chips and the project name are plain text. The only way to a task's page was the complete form's redirect.

## Options

- **Link when there's exactly one page.** If exactly one page loads the entity `by id`, a shown relation links to it. Nothing new to write. When there are two such pages, it doesn't link, which might surprise someone.
- **Declare it.** `shows: project.tasks as chips links_to "/tasks/[id]"`. Exact and checkable, but long, and most specs would write the same thing every time.
- **Never.** Links come only from `links_to` on a component. Records reached through relations stay dead ends.

The first is deterministic and checkable from the spec: the checker already knows every page and what it loads. I'd pick it, and let `tree` print which relations link where so a reviewer can see it.
