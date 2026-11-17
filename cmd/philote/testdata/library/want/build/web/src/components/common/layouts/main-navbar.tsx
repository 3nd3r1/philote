import Link from "next/link";

import { siteConfig } from "@/lib/config";
import { nav } from "@/lib/nav";
import { cn } from "@/lib/utils";

import { buttonVariants } from "@/components/ui/button";

import ModeToggle from "@/components/common/layouts/mode-toggle";

const MainNavbar = () => {
    return (
        <header className="sticky top-0 z-40 w-full border-b bg-background/80 backdrop-blur">
            <div className="mx-auto flex max-w-5xl flex-wrap items-center gap-x-4 px-4 py-2 sm:px-6">
                <Link
                    href="/"
                    className="font-heading text-lg font-semibold tracking-tight"
                >
                    {siteConfig.name}
                </Link>
                {/* On a phone the links get a row of their own, scrolling sideways
                    without a visible bar. */}
                <nav className="order-last -mx-2 flex w-full [scrollbar-width:none] items-center gap-1 overflow-x-auto sm:order-none sm:mx-0 sm:w-auto sm:flex-1">
                    {nav.map((item) => (
                        <Link
                            key={item.href}
                            href={item.href}
                            className={cn(
                                buttonVariants({
                                    variant: "ghost",
                                    size: "sm",
                                }),
                                "text-muted-foreground hover:text-foreground"
                            )}
                        >
                            {item.title}
                        </Link>
                    ))}
                </nav>
                <div className="ml-auto sm:ml-0">
                    <ModeToggle />
                </div>
            </div>
        </header>
    );
};

export default MainNavbar;
