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
  finished_at: datetime optional readonly   // only steps set it
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

Modifiers go after the type: `optional`, `unique`, `readonly`. A readonly field is left out of every resource's input and form, so only steps write it. It must be optional, and some step must list its entity in `writes`. Every entity gets an `id`. Relations are a sentence each, outside the entity.

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

A GET endpoint reads. Its inputs are records named in its route, sent as their ids, and it takes no body. A record that doesn't exist is `not_found`. Under a resource's path, a route looks like `/projects/[project]/open` and can't take one of the resource's own routes, like `/projects/[id]/tasks`.

```
endpoint open_tasks {
  route: GET "/projects/[project]/open"
  input: { project: Project }
  output: list of Task          // a step returns a list of Task
  steps: find_open
}

step find_open(project: Project) -> tasks: list of Task {
  reads: Task
  intent: "the project's tasks that aren't done, oldest first"
}
```

A step's parameters come from the endpoint's input, or from earlier steps' results, by name. The model writes only the step's body. It sees the signature, the intent, and the types of the tables the step can reach. It can only touch the tables the step lists, and the body has to typecheck.

## Frontend

```
frontend web: nextjs {
  uses: api
  look: "calm and focused; deep green, rounded cards"   // optional, up to 200 chars
  nav: bottom, "/tasks" as "Tasks", "/projects"           // optional
  viewer: Member                                          // optional
  ...
}
```

The look sets the whole app's theme: colors for light and dark, which of the two the app opens in, fonts and corner radius. The model writes the theme from it, the compiler checks that every text color is readable, and the theme is pinned in `philote.lock` as `web/theme`. Changing the look rewrites only the theme. Without a look the app gets a plain default theme and no model call.

The nav sits at the `top`, `bottom` or `side` and links to pages without `[id]`, in order. An item's text is its `as` label, up to 24 characters, or the page's title. The template keeps the nav in its place and finds the active item; the model draws the nav from the look, pinned as `web/nav`. Changing an item rewrites only the nav. Without a nav line the app gets a plain top bar with every page that has no `[id]`, other than `/`, then the forms, and no model call.

The viewer is the record a visitor acts as, until there's auth: they pick one, and the app remembers it in a cookie. `C(viewer)` shows it on a page. A component that takes a `list of` the viewer's entity picks it with `picks: viewer then "<route>"`: tapping a record makes it the viewer and opens the route. Exactly one page shows such a component, and a page that shows the viewer sends a visitor who hasn't picked one there first. The backend needs a resource that gets the viewer's entity by id.

```
page "/pick" {
  load: members
  shows: MemberPicker(members)
  empty: "No members yet."
}

page "/me" {
  shows: MemberView(viewer)
}

component MemberPicker(members: list of Member) {
  shows: members.name
  picks: viewer then "/tasks"
}
```

A page loads data and shows components, in order. Each component is a full-width section.

```
page "/" {
  load: tasks                     // a resource's list
  shows: Welcome, TaskBoard(tasks)
  empty: "No tasks yet."
}

page "/tasks" {
  load: tasks
  shows: list of TaskCard(tasks)  // one TaskCard per task, in a grid or in rows
  empty: "No tasks yet."
}

page "/tasks/[id]" {
  load: task from tasks by id     // one record, by the route's [id]
  shows: TaskView(task)
  not_found: "That task doesn't exist."
}

page "/tasks/new" {
  form: tasks.create              // or tasks.update by id, or a custom endpoint
  then: "/tasks/[id]"             // [id] is the saved record
}

component Welcome {
  says: "Get it done.", "Everything your team is working on."
  links_to: "/tasks/new" as "New task"
}

component TaskBoard(tasks: list of Task) {
  shows: tasks.title, tasks.status
  look: "a kanban board, one column per status"
}

component TaskCard(task: Task) {
  shows: task.title, task.status, task.project, task.tags as chips
  links_to: "/tasks/[task.id]"
  look: "compact, status as a colored dot"
  actions: open "/tasks/[task.id]/edit" as "Edit", delete then "/tasks"
}
```

- `load: <name> [from <resource>] [by id]` names what the page fetches. A page loads at most one thing; a page that loads a list says what shows when it's `empty`, a page by id says what shows when it's `not_found`.
- `load: <name> from <endpoint>(<input>: viewer)` loads the list a GET endpoint returns, with the viewer as every input, like `load: mine from open_tasks(project: viewer)`. The page needs the viewer.
- `shows` lists the page's sections. `C(name)` hands a component what the page loads, as it is; `list of C(name)` shows one `C` per record, in a grid or in rows as the component chooses. A component without a parameter, like `Welcome`, takes nothing.
- A component takes no record, one record, or `list of` an entity. Its `shows` lists every field it reads. A belongs-to relation shows the other record's display field. A has-many relation needs `as chips`.
- `says` gives a component its words, up to 200 characters each. The body shows each one exactly as written, and other than these, data and labels, the model writes no text.
- `links_to` makes a card a link to a record's page. With `as "<label>"` it's a button to any page instead. A component that takes a list can't have `links_to` or `actions`.
- The model writes the component's body from its `shows`, `says` and `look`, and places the link and actions the template hands it. A record shown by name gets the same picture, its first letter on its own tint, in every component.
- A shown relation links to its record's page when exactly one page loads that entity `by id`, unless the component has `links_to`. `philote tree` lists these links.
- Form fields are the endpoint's input, in order.

A component can call one custom endpoint that changes something. It then runs in the browser: the model can give it state and gestures, and the template hands its body the endpoint as an action.

```
component SwipeDeck(deck: list of Horse) {
  shows: deck.name, deck.bio
  calls: swipe(horse: viewer, target: deck, liked)
  look: "one card at a time, drag right to like, left to pass"
}
```

- Every input is listed once. `<input>: viewer` is filled on the server with the viewer, so the body never sends who the visitor is, and a page that shows the component needs the viewer. `<input>: <param>` is the id of the component's record, or of the one in its list the visitor acts on. A bare `<input>` is a value from what the visitor does, like a bool.
- The body calls the action, like `await swipe({ target, liked })`, and gets an error message to show, or nothing when it worked. A body that never calls it, or that uses `fetch`, goes back to the model.

## Build

```sh
philote check app.philote
philote tree app.philote             # what the spec becomes
philote expand app.philote api/tasks  # one level down
philote build app.philote            # build/api and build/web; new bodies call the claude CLI
philote build --frozen app.philote   # never calls a model
philote build --redo web/TaskCard app.philote   # or web/theme, web/nav
```

Model-written bodies are pinned in `philote.lock` next to the spec. Commit it. A body is written again only when its prompt changes, and then it starts from the old body.
