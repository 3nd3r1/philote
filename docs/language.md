# The philote language

Everything a spec can say, on one page. Comments start with `//`. `philote check` points at the line when something is wrong.

## App and parts

```
app board {
  backend api: hono {      // a body here, or in api.philote next to this file
    ...
  }
  frontend web: nextjs     // no body: read from web.philote
}
```

A backend uses the `hono` target, and a frontend uses `nextjs`. A frontend starts with `uses: <backend>`.

## Data

```
enum Status { todo, doing, done }

entity Task {
  display: title                 // the field shown when another record shows a Task
  title: string(max 120)
  status: Status
  due: date optional
  code: slug from title unique
}

Project has many tasks: Task                   // Task gets project_id
Task belongs to owner: Person (back owned)     // Person gets owned, a list
```

| Type | Notes |
| --- | --- |
| `string(max N)`, `string(min N, max N)` | |
| `int(min N, max N)` | limits are optional |
| `bool`, `date`, `datetime`, `email`, `markdown` | |
| `slug from <field>` | made from a string field on create |
| an enum's name | |

Modifiers go after the type: `optional`, `unique`. Every entity gets an `id`. Relations are a sentence each, outside the entity.

## Backend

```
resource tasks: Task                        // list, get, create, update, delete
resource swipes: Swipe { only: list, create }
```

A resource with a has-many relation also gets a read route for it, like `GET /projects/[id]/tasks`. `delete` refuses with `in_use` while anything points at the record.

A custom endpoint runs steps in order:

```
endpoint complete {
  route: POST "/complete"
  input: { task: Task, note: string(max 200) }   // a record input is sent as its id
  output: Task
  errors: already_done
  steps: finish_task
}

step finish_task(task: Task) -> done: Task {
  fails: already_done
  reads: Project       // tables it only reads
  writes: Task         // tables it changes, and may read
  intent: "mark the task done and set finished_at to now; fail if it's already done"
}
```

A step's parameters come from the endpoint's input, or from earlier steps' results, by name. The model writes only the step's body. It sees the signature, the intent, and the types of the tables the step can reach. It can only touch the tables the step lists, and the body has to typecheck.

## Frontend

```
page "/tasks" {
  data: tasks
  layout: list of TaskCard
  empty: "No tasks yet."
}

page "/tasks/[id]" {
  data: tasks by id
  layout: TaskView
  not_found: "That task doesn't exist."
}

page "/tasks/new" {
  form: tasks.create              // or tasks.update by id, or a custom endpoint
  then: "/tasks/[id]"             // [id] is the saved record
}

component TaskCard(task: Task) {
  shows: task.title, task.status, task.project, task.tags as chips
  links_to: "/tasks/[task.id]"
  look: "compact, status as a colored dot"
  actions: open "/tasks/[task.id]/edit" as "Edit", delete then "/tasks"
}
```

- `shows` lists every field the component reads. A belongs-to relation shows the other record's display field. A has-many relation needs `as chips`.
- The model writes the component's body from its `shows` and `look`. Links and actions are rendered around the body by templates.
- Form fields are the endpoint's input, in order. Pages without `[id]` go in the nav.

## Build

```sh
philote check app.philote
philote tree app.philote             # what the spec becomes
philote expand app.philote api/tasks  # one level down
philote build app.philote            # build/api and build/web; new bodies call the claude CLI
philote build --frozen app.philote   # never calls a model
philote build --redo web/TaskCard app.philote
```

Model-written bodies are pinned in `philote.lock` next to the spec. Commit it. A body is written again only when its prompt changes, and then it starts from the old body.
