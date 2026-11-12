import type { Issue } from "@/lib/backend";

// FormState is what a form's server action returns when a save fails: the
// error's message, each field's issue, and what was typed, so the form can
// show it again.
export type FormState = {
    error?: string;
    issues?: Record<string, string>;
    values?: Record<string, string>;
};

// Messages for the built-in errors. A custom error shows as its name in
// words (D-020).
const messages: Record<string, string> = {
    invalid_input: "Some fields need another look.",
    not_found: "That record doesn't exist anymore.",
    in_use: "Other records still point at this one, so it can't be deleted.",
};

export function message(error: string): string {
    const words = error.replaceAll("_", " ");
    return messages[error] ?? words[0].toUpperCase() + words.slice(1) + ".";
}

// The readers below return what a valid form sends. An empty field is sent
// as missing and a bad number as typed, so the backend, which checks every
// input, reports it against the field.

export function text<T extends string = string>(
    data: FormData,
    name: string
): T {
    const value = data.get(name);
    return (typeof value === "string" && value !== "" ? value : undefined) as T;
}

export function int(data: FormData, name: string): number {
    const value = text(data, name);
    const n = Number(value);
    return (value === undefined || Number.isNaN(n) ? value : n) as number;
}

export function bool(data: FormData, name: string): boolean {
    return data.get(name) === "on";
}

// dateTime reads a datetime-local field as the ISO string the backend takes.
export function dateTime(data: FormData, name: string): string {
    const value = text(data, name);
    const date = new Date(value);
    return value === undefined || Number.isNaN(date.getTime())
        ? value
        : date.toISOString();
}

// localDateTime is an ISO string as a datetime-local field's value.
export function localDateTime(iso: string): string {
    const d = new Date(iso);
    const pad = (n: number) => String(n).padStart(2, "0");
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// failed is the state a form shows after the backend refused a save.
export function failed(
    data: FormData,
    error: string,
    issues: Issue[] | undefined,
    notFound?: string
): FormState {
    const byField: Record<string, string> = {};
    for (const issue of issues ?? []) {
        byField[issue.field] ??= issue.message;
    }
    const values: Record<string, string> = {};
    data.forEach((value, name) => {
        if (typeof value === "string" && !name.startsWith("$ACTION")) {
            values[name] = value;
        }
    });
    const text = error === "not_found" && notFound ? notFound : message(error);
    return { error: text, issues: byField, values };
}
