import { z } from "zod";

export const siteConfigSchema = z.object({
    name: z.string().min(1),
    shortName: z.string().min(1),
    description: z.string().min(1),
    url: z.string(),
    locale: z.string().min(1),
    language: z.string().min(1),
    author: z.object({
        name: z.string().min(1),
        url: z.string(),
    }),
    keywords: z.array(z.string()),
    category: z.string().min(1),
});

export type SiteConfig = z.infer<typeof siteConfigSchema>;
