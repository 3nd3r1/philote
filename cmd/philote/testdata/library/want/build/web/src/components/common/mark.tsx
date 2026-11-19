import { cn } from "@/lib/utils";

// hues turn the theme's primary a little, so records that sit side by side
// tell apart while every tint still belongs to the theme.
const hues = [-50, -25, 0, 25, 50];

// Mark is a record's picture: the first letter of its name on a tint. The
// tint comes from the name alone, so one record looks the same in every
// component that shows it. The caller sizes and shapes it with className.
export default function Mark({
    name,
    className,
}: {
    name: string;
    className?: string;
}) {
    let hash = 0;
    for (const ch of name) {
        hash = (hash * 31 + ch.codePointAt(0)!) >>> 0;
    }
    const tint = `oklch(from var(--primary) l c calc(h + ${hues[hash % hues.length]}))`;
    return (
        <span
            aria-hidden="true"
            className={cn(
                "[container-type:size] flex aspect-square shrink-0 items-center justify-center overflow-hidden rounded-full select-none",
                className
            )}
            style={{
                backgroundImage: `radial-gradient(circle at 30% 20%, color-mix(in oklch, ${tint} 10%, var(--card)), color-mix(in oklch, ${tint} 32%, var(--card)))`,
                color: `color-mix(in oklch, ${tint} 85%, var(--card-foreground))`,
            }}
        >
            <span className="font-heading text-[45cqmin] leading-none font-semibold">
                {Array.from(name.trim())[0]?.toUpperCase() ?? ""}
            </span>
        </span>
    );
}
