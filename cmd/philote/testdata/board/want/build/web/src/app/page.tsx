import { redirect } from "next/navigation";

import { siteConfig } from "@/lib/config";
import { nav } from "@/lib/nav";

// The spec names no home page, so the app opens on its first page.
export default function Home() {
    if (nav.length > 0) {
        redirect(nav[0].href);
    }
    return (
        <main className="flex-1 px-4 py-10">
            <div className="mx-auto max-w-4xl space-y-2 text-center">
                <h1 className="font-heading text-4xl">{siteConfig.name}</h1>
                <p className="text-muted-foreground">
                    {siteConfig.description}
                </p>
            </div>
        </main>
    );
}
