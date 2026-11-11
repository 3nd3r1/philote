+++
id = "D-016"
title = "Frontend first slice: a Next.js project with a typed client"
status = "accepted"
summary = "A frontend part names the backend it uses and builds a runnable Next.js project, based on the user's template, with a generated, typed, server-only client for that backend. No pages and no model yet."
created = 2026-11-10
related = ["D-013", "D-014", "D-001", "Q-003", "Q-005", "Q-012"]
+++

## Decision

A frontend part has one line, `uses`, naming a backend part in the same app:

```
app neighbr {
  backend api: hono
  frontend web: nextjs {
    uses: api
  }
}
```

`philote build` writes `build/web/`: a base project plus generated files.

The base is a pinned copy of the user's Next.js template (`3nd3r1/nextjs-template`): app router, Tailwind, shadcn/ui on Base UI, site config in `src/lib/config.ts`, theme in `globals.css`, metadata files, ESLint and Prettier. Exact versions in `package.json`, like the Hono base. Left out: the showcase content (about page, cards, ping route), tests, Docker and CI. The home page shows the app's name.

Generated, both in `src/lib/` where the template keeps shared code:

- `types.ts`: the enums and entities, the same shapes the backend returns.
- `api.ts`: one function per endpoint of the used backend, resources and custom endpoints alike. Inputs are named types (`NewHorse` for a resource, `SwipeInput` for a custom endpoint), typed the way D-014 says: entity fields sent as ids. Each function returns a result instead of throwing, with the endpoint's errors as a union:

```ts
export const horses = {
    list: () =>
        call<Horse[], never>("GET", "/horses", []),
    get: (id: string) =>
        call<Horse, "not_found">("GET", `/horses/${encodeURIComponent(id)}`, ["not_found"]),
    ...
};

export const swipe = (input: SwipeInput) =>
    call<Swipe, "invalid_input" | "same_horse" | "already_swiped">("POST", "/swipe", [...], input);
```

`call` returns `Result<T, E>`: `{ ok: true; value: T }` or `{ ok: false; error: E; issues?: Issue[] }`.

The error unions come from each endpoint's errors in the IR, which `lower` fills for resources and `check` for custom endpoints. Both targets read the same list. Network failures and unexpected statuses still throw. They aren't part of the spec.

`config.ts` is generated too, from the app's name. The request code shared by every call (`call`, `Result`, `Issue`) is a plain file in the base, `src/lib/backend.ts`. Generated files use the template's style but are listed in `.prettierignore`: the compiler owns their layout, and long signatures would otherwise fight Prettier's line breaks.

The client runs on the server only. `api.ts` imports `server-only`, so importing it from a client component fails the build. It reads the backend's address from `API_URL`. So the backend needs no CORS, and no API address ends up in the browser.

`tree` shows the part as `web (frontend, nextjs) uses api`. Check errors:

- `uses` is missing.
- `uses` names something that isn't a backend part (hint: the backend parts).
- A frontend part has any line other than `uses`, or `uses` appears twice.
- `uses` appears in a backend part.

`Verify` runs `next typegen` and `tsc`. The smoke test starts the backend and calls it through the generated client, not only with curl.

## Out of scope

Pages, components, model-written frontend code, browser-side calls, auth. They depend on Q-003, Q-005, Q-011 and Q-012, which stay open. Shipping shadcn and the template's theme tokens in the base leans Q-003 toward tokens, but answers nothing: no generated code uses them yet.

## Why

D-013 promises the frontend a typed client for the backend. That is the one frontend piece every later design needs, whatever pages turn out to be. The client and the backend handlers come from the same endpoint IR, so they can't drift apart: renaming a field or adding an error changes both sides in one build (D-001). It needs no model, so it stays fully deterministic, and it opens no question that pages would have to answer later.

The user's template is the structure they'd write by hand: files grouped by kind, one site config and one theme file that everything reads from. Generated code goes where the template would put it, so a person reading the output finds things where they expect.
