import { siteConfig } from "@/lib/config";

export default function Home() {
    return (
        <main className="flex-1 p-8">
            <div className="mx-auto max-w-4xl space-y-2 text-center">
                <h1 className="font-heading text-4xl">{siteConfig.name}</h1>
                <p className="text-muted-foreground">
                    {siteConfig.description}
                </p>
            </div>
        </main>
    );
}
