+++
id = "D-019"
title = "Components: model-written bodies, pinned like steps"
status = "accepted"
summary = "A component's JSX is written by the model from its props, the base's shadcn components and an optional look line, then pinned in philote.lock. The spec still owns data and links."
created = 2026-11-11
answers = ["Q-017"]
related = ["D-015", "D-017", "D-018", "D-001", "P-007"]
+++

## Decision

Template cards printed every field as a label and a value. The result was correct and looked like a schema dump. A component's body is now written by the model, the same way a step's is (D-015):

1. **The compiler writes the file around the body.** It fixes the imports (the base's shadcn components, lucide icons, the format helpers, `cn`), the props type from `shows` (D-017), and the function signature. If the component has `links_to`, the exported component wraps the body in the link, so the model can't drop or change it.
2. **The model writes only the JSX body.** The prompt gives it the props type, the available components and their variants, the theme's conventions, where the component is used (a list grid or a page of its own), and the component's `look` line if it has one.
3. **The body is typechecked and pinned.** It goes through `next typegen` and `tsc`, back to the model with the errors if it fails, and into `philote.lock` keyed by a hash of the model and the prompt. A rebuild copies it byte for byte with no model call. Changing a component's `shows` or `look` rewrites only that component, starting from its old body.
4. **`look` steers the design.** `look: "the name big, the breed as a badge, the age small and muted"` is optional, capped at 200 characters, and reaches the model like a step's intent.
5. **`--redo web/HorseCard` asks for a fresh body** when you don't like the one that's pinned. It ignores the old body, so you get a new design, not a tweak. It can't be combined with `--frozen`.

Pages stay templates: data loading, the empty and not-found states, the heading and the grid. Data never comes from the model. It writes code, and the backend still holds the only records.

## Why

The user wants the frontend to be able to look like a real product (Tinder, with swipes and a wall of matches) while staying deterministic. Templates are deterministic but can only be as expressive as the layouts written into the compiler. Every new kind of UI would need a new keyword. Pinned model output is deterministic too: the same spec and lock give the same files, and nothing changes without a reviewable diff in the lock.

Each component's prompt is built only from the spec and the target, never from other components' bodies, so a change stays local (P-005). The cost is that components generated separately can drift in style. The prompt's shared theme conventions limit that, and `--redo` fixes one that stands out.

This changes the project rule "the default is a template": a frontend component's body is model-written by default. Backend resources and every page stay templates.
