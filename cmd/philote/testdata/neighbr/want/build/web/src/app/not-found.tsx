import { Separator } from "@/components/ui/separator";

export default function NotFound() {
    return (
        <div className="flex flex-1 items-center justify-center">
            <div className="flex flex-row gap-x-2">
                <h2 className="text-xl">404</h2>
                <Separator orientation="vertical" />
                <div className="flex items-center">
                    <p>That page could not be found</p>
                </div>
            </div>
        </div>
    );
}
