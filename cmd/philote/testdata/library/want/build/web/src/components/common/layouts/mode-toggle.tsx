"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";

import { Button } from "@/components/ui/button";

// Both icons render and CSS picks one, so the server and the first client
// render agree before the theme is known.
const ModeToggle = () => {
    const { resolvedTheme, setTheme } = useTheme();
    return (
        <Button
            variant="ghost"
            size="icon"
            aria-label="Switch between light and dark"
            onClick={() =>
                setTheme(resolvedTheme === "dark" ? "light" : "dark")
            }
        >
            <Sun className="dark:hidden" />
            <Moon className="hidden dark:block" />
        </Button>
    );
};

export default ModeToggle;
