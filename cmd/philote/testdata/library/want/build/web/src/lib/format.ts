// Dates come from the backend as ISO strings. They're shown in UTC, so a
// date reads the same on every server.
export function formatDate(value: string): string {
    return new Date(value).toLocaleDateString("en", {
        dateStyle: "medium",
        timeZone: "UTC",
    });
}

export function formatDateTime(value: string): string {
    return new Date(value).toLocaleString("en", {
        dateStyle: "medium",
        timeStyle: "short",
        timeZone: "UTC",
    });
}

export function formatBool(value: boolean): string {
    return value ? "yes" : "no";
}
