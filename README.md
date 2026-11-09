<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.png">
  <img src="docs/assets/logo.png" alt="philote" width="380">
</picture>

Can we bring Nix-style determinism to AI-driven coding?

When I build web apps and backends with AI, the idea it generates usually makes sense, but each run picks different technologies I have no context on, and the results vary. Philote tries to fix that. You write a strict spec of what the app does. The compiler turns it into a Next.js frontend and a Hono backend, mostly from templates, so the stack is always the same. A model fills only the gaps the templates mark, and its output is pinned in a lockfile the way Nix pins builds, so the same spec always gives the same code.

Status: early design. There is no CLI to install yet.

To work on philote, see [CONTRIBUTING.md](CONTRIBUTING.md).
