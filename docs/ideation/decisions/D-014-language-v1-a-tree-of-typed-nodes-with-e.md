+++
id = "D-014"
title = "Language v1: a tree of typed nodes, with endpoints as objects"
status = "proposed"
summary = "Own syntax with braces and keyword-led lines. The app is a tree of nodes. Data is entities plus relation sentences. Backend endpoints are objects; resources expand to CRUD by template."
created = 2026-11-09
answers = ["Q-001", "Q-015", "Q-016"]
related = ["I-009", "D-001", "D-013", "P-007", "Q-005"]
+++

## Decision

- **Own syntax** (Q-001). Blocks use braces, and every line inside starts with a keyword or a name. A YAML or TOML file would still need a second parser for routes, types and relations, and the prototype showed the parser is the easy part.
- **Readable, not English** (Q-016). Lines read aloud as phrases (`Order has many lines: OrderLine`, `route: POST "/checkout"`), but the grammar is small and fixed. No free sentences.
- **The app is a tree of nodes.** Each node has a path (`api`, `api/orders`, `api/checkout`), which is also its cache key id and where its generated code goes.
- **Files follow the tree.** A node declared without a body is loaded from a file named after it, like `mod api;` in Rust. Small apps stay in one file.
- **Built-in relation verbs are `has many` and `belongs to`** (Q-015). Made-up verbs (I-002) wait for v2.
- **Endpoints are objects:** route, input, output, errors, steps. A `resource` is the zoomed-out form and expands to CRUD endpoints by template.
- **Steps are the only place the model writes code.** Each step has a typed signature, the errors it can raise and a short intent. Everything else comes from templates.

## Example

`shop.philote`:

```
app shop {
  backend api: hono      // no body: loaded from api.philote
  frontend web: nextjs
}

enum OrderStatus { pending, paid, shipped, cancelled }

entity Product {
  name: string(max 120)
  slug: slug from name
  price_cents: int(min 0)
  stock: int(min 0)
}

entity Customer {
  email: email unique
  name: string(max 80)
}

entity Order {
  status: OrderStatus
  placed_at: datetime
}

entity OrderLine {
  quantity: int(min 1)
  price_cents: int(min 0)
}

Customer has many orders: Order
Order has many lines: OrderLine
OrderLine belongs to product: Product
```

`api.philote`, the body of `backend api`:

```
resource products: Product { only: list, get }
resource orders: Order { only: list, get }

type CheckoutLine {
  product: Product         // sent as its id
  quantity: int(min 1)
}

endpoint checkout {
  route: POST "/checkout"
  input: { email: email, lines: [CheckoutLine](min 1) }
  output: Order
  errors: out_of_stock(product: Product), payment_failed
  steps: reserve_stock, create_order, charge_payment, send_receipt
}

step reserve_stock(lines: [CheckoutLine]) {
  fails: out_of_stock
  intent: "lower stock for each line, fail on the first product without enough"
}
```

The other three steps look like `reserve_stock`. `philote expand api/products` prints what the resource becomes:

```
endpoint products.list { route: GET "/products"       output: [Product] }
endpoint products.get  { route: GET "/products/[id]"  output: Product  errors: not_found }
```

## Rules

- Every entity gets `id`. Relations add the foreign key on the `belongs to` side. The reverse name defaults to the other entity's name in lowercase (`order.customer`). If that's ambiguous, it's an error asking for an explicit name.
- In a `type` or `input`, an entity-typed field is sent as that entity's id. In an `output`, it's the full record.
- Route params use `[name]` everywhere. Targets translate them (`:name` in Hono).
- Step params are bound by name from the endpoint input and earlier steps' outputs. The checker verifies each one exists with the right type. Exactly one step must return the endpoint's output type.
- A step can only raise errors its endpoint declares. Every declared error must be raisable by some step or by the template (like `not_found`).
- `intent` is the one natural-language field the model sees. It's capped at 200 characters (the exception P-007 mentions).

## First slice

Build only the backend with resources: `philote check`, `philote tree` and `philote build` on `shop.philote`, generating a runnable Hono server from templates. No model calls and no steps. Data lives in a JSON file (D-009). Golden tests cover the output. This answers whether endpoints-as-objects give a deterministic backend.

Then steps with the model, then the Next.js frontend reading through a generated client.

## Out of scope for v1

Auth, made-up relation verbs, many-to-many, migrations, pagination, file uploads.

## Why

Each piece maps to a node with an interface the compiler owns (D-001), so templates cover everything except steps (D-013). The tree gives zoom (I-009) and file layout from one rule. Relation sentences make the data model readable at a glance, which is what review needs (P-008).
