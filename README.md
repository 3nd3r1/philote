<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.png">
  <img src="docs/assets/logo.png" alt="philote" width="380">
</picture>

Can we bring Nix-style determinism to AI-driven coding?

Describe your app as an ontology: the things in it, how they relate, and what you can do with them. Philote compiles it to code, from fixed templates wherever it can. Anything AI writes is pinned in a lockfile, so the same ontology always builds the same code.

Status: early. It builds a Hono backend. AI-written steps and a Next.js frontend are next.

## Try it

```sh
go install github.com/3nd3r1/philote/cmd/philote@latest
```

`shop.philote`:

```
app shop {
  backend api: hono {
    resource products: Product
  }
}

entity Product {
  name: string(max 120)
  slug: slug from name
  price_cents: int(min 0)
}
```

```sh
philote check shop.philote             # errors point at the spec line
philote tree --depth 3 shop.philote    # the app as a tree, down to routes
philote expand shop.philote api        # what each resource becomes, as spec
philote build shop.philote             # writes a runnable server to build/api
cd build/api && npm install && npm start
```

To work on philote, see [CONTRIBUTING.md](CONTRIBUTING.md).
