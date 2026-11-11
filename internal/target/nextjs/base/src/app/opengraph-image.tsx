import { ImageResponse } from "next/og";

import { siteConfig } from "@/lib/config";
import { getThemeColors } from "@/lib/theme";

export const size = {
    width: 1200,
    height: 630,
};

export const contentType = "image/png";

export default function Image() {
    const colors = getThemeColors();

    return new ImageResponse(
        <div
            style={{
                fontSize: 64,
                background: colors.primary,
                color: colors.primaryForeground,
                width: "100%",
                height: "100%",
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                justifyContent: "center",
                padding: 48,
            }}
        >
            <div style={{ fontSize: 72, fontWeight: 700 }}>
                {siteConfig.name}
            </div>
            <div
                style={{
                    fontSize: 32,
                    marginTop: 24,
                    opacity: 0.8,
                    textAlign: "center",
                }}
            >
                {siteConfig.description}
            </div>
        </div>,
        { ...size }
    );
}
