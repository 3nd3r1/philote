import Link from "next/link";

import { Library } from "lucide-react";

import { siteConfig } from "@/lib/config";
import { nav } from "@/lib/nav";
import { cn } from "@/lib/utils";

import { buttonVariants } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

const MainNavbar = () => {
    return (
        <header className="w-full border-b">
            <div className="flex flex-row">
                <div className="px-4 py-3">
                    <Library />
                </div>
                <Separator orientation="vertical" />
                <nav className="grow px-4 py-2">
                    <Link
                        href="/"
                        className={cn(buttonVariants({ variant: "ghost" }))}
                    >
                        {siteConfig.name}
                    </Link>
                    {nav.map((item) => (
                        <Link
                            key={item.href}
                            href={item.href}
                            className={cn(buttonVariants({ variant: "ghost" }))}
                        >
                            {item.title}
                        </Link>
                    ))}
                </nav>
            </div>
        </header>
    );
};

export default MainNavbar;
