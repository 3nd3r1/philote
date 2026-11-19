import type { MetadataRoute } from "next";

import { siteConfig } from "@/lib/config";
import { getThemeColors } from "@/lib/theme";

export default function manifest(): MetadataRoute.Manifest {
    const colors = getThemeColors();

    return {
        name: siteConfig.name,
        short_name: siteConfig.shortName,
        description: siteConfig.description,
        start_url: "/",
        display: "standalone",
        background_color: colors.background,
        theme_color: colors.primary,
        icons: [
            {
                src: "/favicon.ico",
                sizes: "48x48",
                type: "image/x-icon",
            },
            {
                src: "/android-chrome-192x192.png",
                sizes: "192x192",
                type: "image/png",
            },
            {
                src: "/android-chrome-512x512.png",
                sizes: "512x512",
                type: "image/png",
            },
        ],
    };
}
