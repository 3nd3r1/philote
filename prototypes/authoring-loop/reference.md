# Spec language v0

One `app` block, then entities, pages and components in any order. Every field is one line. `//` starts a comment.

## app

    app <name> {
      target: nextjs
      style: tokens/minimal
    }

## entity

    entity <Name> {
      <field>: <type> [unique] [optional]
    }

Types: `string`, `string(max N)`, `slug from <string field>`, `markdown`, `date`, `int`, `bool`, another entity `<Entity>`, or a list `[<Entity>]` / `[string]`.
`id` is automatic and reserved. Slugs count as unique.

## page

    page "<route>" {
      data: <query>
      layout: <Component>            // for a single record
      layout: list of <Component>    // for a list
      empty: "<text>"                // required on list pages, max 200 chars
      not_found: 404                 // required on single-record pages
    }

Queries:

- `<Entity> by <field>`: one record. The field must be unique, and the route must have exactly that param, e.g. `"/people/[handle]"` with `Person by handle`.
- `<Entity> [where <field> <op> <value> [and ...]] [order by <field> [asc|desc]] [limit N]`: a list. List pages can't have route params.
  Ops: `== != < <= > >=`. Values: `now`, `true`, `false`, a number, a `"string"`.

## component

    component <Name>(<param>: <Type>) {
      shows: <item>, <item>, ...
      links_to: "/route/[param.field]"
    }

A layout component takes exactly one parameter, of the page's entity.

Items in `shows`:

- a field path: `person.name`, `post.author.name`
- `excerpt(<markdown or string>, N)`
- `<markdown field> as markdown`. Markdown can't be shown any other way.
- `<list> as chips`. For a list of records, the record must have exactly one `string` field.

`links_to` must match a page's route, and pass the field that page looks up by.

Nothing else exists: no forms or actions, no search, no other functions or renderers.
