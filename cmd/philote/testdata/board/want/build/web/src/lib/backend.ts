import "server-only";

export type Issue = { field: string; message: string };

export type Result<T, E extends string> =
    { ok: true; value: T } | { ok: false; error: E; issues?: Issue[] };

// call sends one request to the backend at API_URL. Errors the endpoint
// declares come back as results. Anything else throws: the spec doesn't
// describe it, so no caller can handle it.
export async function call<T, E extends string>(
    method: string,
    path: string,
    errors: readonly E[],
    body?: unknown
): Promise<Result<T, E>> {
    const base = process.env.API_URL;
    if (!base) {
        throw new Error("API_URL is not set");
    }
    const res = await fetch(base.replace(/\/$/, "") + path, {
        method,
        headers: { "content-type": "application/json" },
        body: body === undefined ? undefined : JSON.stringify(body),
    });
    if (res.ok) {
        const value = res.status === 204 ? undefined : await res.json();
        return { ok: true, value: value as T };
    }
    const data = await res.json().catch(() => undefined);
    if (errors.includes(data?.error)) {
        return { ok: false, error: data.error, issues: data.issues };
    }
    throw new Error(`${method} ${path}: unexpected status ${res.status}`);
}

// must unwraps a result a page can't recover from, like a record that was
// deleted between two reads.
export async function must<T>(result: Promise<Result<T, string>>): Promise<T> {
    const r = await result;
    if (!r.ok) {
        throw new Error(`the backend returned ${r.error}`);
    }
    return r.value;
}
