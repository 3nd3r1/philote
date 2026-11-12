import { Label } from "@/components/ui/label";

// FormField is one labelled input with the issue the backend found in it.
export default function FormField({
    name,
    label,
    issue,
    children,
}: {
    name: string;
    label: string;
    issue?: string;
    children: React.ReactNode;
}) {
    return (
        <div className="space-y-2">
            <Label htmlFor={name}>{label}</Label>
            {children}
            {issue && <p className="text-xs text-destructive">{issue}</p>}
        </div>
    );
}
