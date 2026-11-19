"use client";

import { useEffect, useState } from "react";

import { cn } from "@/lib/utils";

// Counter shows how many characters the field with this id holds, out of
// max. It follows the field's input events, so the field stays uncontrolled.
export default function Counter({
    id,
    max,
    initial,
}: {
    id: string;
    max: number;
    initial: number;
}) {
    const [count, setCount] = useState(initial);
    useEffect(() => {
        const field = document.getElementById(id) as
            HTMLInputElement | HTMLTextAreaElement | null;
        if (!field) {
            return;
        }
        const update = () => {
            setCount(field.value.length);
        };
        field.addEventListener("input", update);
        return () => {
            field.removeEventListener("input", update);
        };
    }, [id]);
    return (
        <p
            aria-live="polite"
            className={cn(
                "text-right text-xs text-muted-foreground tabular-nums",
                count > max * 0.9 && "text-foreground",
                count > max && "text-destructive"
            )}
        >
            {count}/{max}
        </p>
    );
}
