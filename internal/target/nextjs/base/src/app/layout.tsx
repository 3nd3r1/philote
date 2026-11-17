import type { Metadata, Viewport } from "next";
import {
    Bricolage_Grotesque,
    DM_Sans,
    DM_Serif_Display,
    Fraunces,
    Geist,
    Geist_Mono,
    Instrument_Serif,
    Inter,
    Manrope,
    Nunito,
    Playfair_Display,
    Space_Grotesk,
} from "next/font/google";

import { siteConfig } from "@/lib/config";
import { getThemeColors } from "@/lib/theme";
import { cn } from "@/lib/utils";

import { Toaster } from "@/components/ui/sonner";

import MainNavbar from "@/components/common/layouts/main-navbar";
import ThemeProvider from "@/components/common/layouts/theme-provider";

import "./globals.css";

// Every font a theme can pick, as --font-<name>. next/font needs one const
// per font. None is preloaded, so a browser downloads only the fonts the
// theme's CSS uses.
const geist = Geist({
    variable: "--font-geist",
    subsets: ["latin"],
    preload: false,
});

const inter = Inter({
    variable: "--font-inter",
    subsets: ["latin"],
    preload: false,
});

const dmSans = DM_Sans({
    variable: "--font-dm-sans",
    subsets: ["latin"],
    preload: false,
});

const manrope = Manrope({
    variable: "--font-manrope",
    subsets: ["latin"],
    preload: false,
});

const nunito = Nunito({
    variable: "--font-nunito",
    subsets: ["latin"],
    preload: false,
});

const spaceGrotesk = Space_Grotesk({
    variable: "--font-space-grotesk",
    subsets: ["latin"],
    preload: false,
});

const bricolageGrotesque = Bricolage_Grotesque({
    variable: "--font-bricolage-grotesque",
    subsets: ["latin"],
    preload: false,
});

const fraunces = Fraunces({
    variable: "--font-fraunces",
    subsets: ["latin"],
    preload: false,
});

const playfairDisplay = Playfair_Display({
    variable: "--font-playfair-display",
    subsets: ["latin"],
    preload: false,
});

const dmSerifDisplay = DM_Serif_Display({
    variable: "--font-dm-serif-display",
    weight: "400",
    subsets: ["latin"],
    preload: false,
});

const instrumentSerif = Instrument_Serif({
    variable: "--font-instrument-serif",
    weight: "400",
    subsets: ["latin"],
    preload: false,
});

const geistMono = Geist_Mono({
    variable: "--font-mono",
    subsets: ["latin"],
    preload: false,
});

const fonts = [
    geist,
    inter,
    dmSans,
    manrope,
    nunito,
    spaceGrotesk,
    bricolageGrotesque,
    fraunces,
    playfairDisplay,
    dmSerifDisplay,
    instrumentSerif,
    geistMono,
];

export const metadata: Metadata = {
    metadataBase: new URL(siteConfig.url),
    title: {
        default: siteConfig.name,
        template: `%s | ${siteConfig.name}`,
    },
    description: siteConfig.description,
    keywords: siteConfig.keywords,
    authors: [{ name: siteConfig.author.name, url: siteConfig.author.url }],
    creator: siteConfig.author.name,
    robots: "index, follow",
    openGraph: {
        type: "website",
        locale: siteConfig.locale,
        url: siteConfig.url,
        siteName: siteConfig.name,
        title: siteConfig.name,
        description: siteConfig.description,
    },
    twitter: {
        card: "summary_large_image",
        title: siteConfig.name,
        description: siteConfig.description,
    },
    icons: {
        icon: "/favicon.ico",
        shortcut: "/favicon.ico",
        apple: "/apple-touch-icon.png",
    },
    category: siteConfig.category,
};

export const viewport: Viewport = {
    themeColor: getThemeColors().primary,
};

const structuredData = {
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: siteConfig.name,
    description: siteConfig.description,
    url: siteConfig.url,
    author: {
        "@type": "Person",
        name: siteConfig.author.name,
        url: siteConfig.author.url,
    },
};

export default function RootLayout({ children }: LayoutProps<"/">) {
    return (
        <html
            lang={siteConfig.language}
            className={cn(
                ...fonts.map((font) => font.variable),
                "h-full font-sans antialiased"
            )}
            suppressHydrationWarning
        >
            <head>
                {/* The app has its own dark mode, so Dark Reader stays off it. */}
                <meta name="darkreader-lock" />
                <script
                    type="application/ld+json"
                    dangerouslySetInnerHTML={{
                        __html: JSON.stringify(structuredData),
                    }}
                />
            </head>
            <body className="flex min-h-full flex-col">
                <ThemeProvider
                    attribute="class"
                    defaultTheme="system"
                    enableSystem
                    disableTransitionOnChange
                >
                    <div className="flex flex-1 flex-col">
                        <MainNavbar />
                        {children}
                    </div>
                    <Toaster />
                </ThemeProvider>
            </body>
        </html>
    );
}
