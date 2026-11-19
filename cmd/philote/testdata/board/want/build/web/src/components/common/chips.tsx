import { cn } from "@/lib/utils";

// Chips picks one of a few values. Each chip is a radio input styled as a
// pill, so it posts with the form and needs no JavaScript.
export default function Chips({
    id,
    name,
    choices,
    defaultValue,
    required,
    invalid,
}: {
    id: string;
    name: string;
    choices: { value: string; label: string }[];
    defaultValue?: string;
    required?: boolean;
    invalid?: boolean;
}) {
    return (
        <div
            id={id}
            role="radiogroup"
            aria-invalid={invalid}
            className="flex flex-wrap gap-2"
        >
            {choices.map((c) => (
                <label
                    key={c.value}
                    className={cn(
                        "cursor-pointer rounded-full border border-input px-3.5 py-1.5 text-sm font-medium shadow-xs transition-colors select-none hover:bg-accent hover:text-accent-foreground",
                        "has-checked:border-primary has-checked:bg-primary has-checked:text-primary-foreground",
                        "has-focus-visible:ring-3 has-focus-visible:ring-ring/50",
                        invalid && "border-destructive"
                    )}
                >
                    <input
                        type="radio"
                        name={name}
                        value={c.value}
                        defaultChecked={defaultValue === c.value}
                        required={required}
                        className="sr-only"
                    />
                    {c.label}
                </label>
            ))}
        </div>
    );
}
