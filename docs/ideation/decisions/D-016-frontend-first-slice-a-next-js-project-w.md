+++
id = "D-016"
title = "Frontend first slice: a Next.js project with a typed client"
status = "proposed"
summary = "A frontend part names the backend it uses and builds a runnable Next.js project with a generated, typed client for that backend. No pages and no model yet."
created = 2026-11-10
related = ["D-013", "D-014", "D-001", "Q-005", "Q-012"]
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

`philote build` writes `build/web/`: a pinned Next.js base project (app router, one static home page) plus two generated files.

- `src/types.ts`: the enums and entities, the same shapes the backend returns.
- `src/api.ts`: one function per endpoint of the used backend, resources and custom endpoints alike. Inputs are typed the way D-014 says (entity fields sent as ids). Each function returns a result instead of throwing, with the endpoint's declared errors as a union:

```ts
export type Result<T, E extends string> =
  | { ok: true; value: T }
  | { ok: false; error: E; issues?: Issue[] };

export const horses = {
  list: (): Promise<Result<Horse[], never>> => ...,
  get: (id: string): Promise<Result<Horse, "not_found">> => ...,
  ...
};

export function swipe(input: { horse: string; target: string; liked: boolean }):
  Promise<Result<Swipe, "invalid_input" | "same_horse" | "already_swiped">>
```

Network failures and unexpected statuses still throw. They aren't part of the spec.

The client runs on the server only (server components, route handlers, server actions) and reads the backend's address from `API_URL`. So the backend needs no CORS, and no API address ends up in the browser.

`tree` shows the part as `web (frontend, nextjs) uses api`. Check errors:

- `uses` is missing.
- `uses` names something that isn't a backend part (hint: the closest part name).
- A frontend part has any line other than `uses`.

`Verify` typechecks the project with `tsc`. The smoke test starts the backend and calls it through the generated client, not only with curl.

## Out of scope

Pages, components, model-written frontend code, browser-side calls, auth. They depend on Q-003, Q-005, Q-011 and Q-012, which stay open.

## Why

D-013 promises the frontend a typed client for the backend. That is the one frontend piece every later design needs, whatever pages turn out to be. The client and the backend handlers come from the same endpoint IR, so they can't drift apart: renaming a field or adding an error changes both sides in one build (D-001). It needs no model, so it stays fully deterministic, and it opens no question that pages would have to answer later.
