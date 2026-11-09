<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.png">
  <img src="docs/assets/logo.png" alt="philote" width="380">
</picture>

Can we bring Nix-style determinism to AI-driven coding?

Describe your app as an ontology: the things in it, how they relate, and what you can do with them. Philote compiles it to code, from fixed templates wherever it can. Anything AI writes is pinned in a lockfile, so the same ontology always builds the same code.

Status: early. It builds a Hono backend. AI-written steps and a Next.js frontend are next.

## Example

You write what the app is: its data, how the data relates, and which parts the API exposes.

```
app shop {
  backend api: hono {
    resource products: Product
    resource customers: Customer { only: list, get, create }
    resource orders: Order { only: list, get }
  }
}

enum OrderStatus { pending, paid, shipped }

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

Customer has many orders: Order
```

`philote build shop.philote` turns that into a runnable Hono server. Every route, rule and error comes from a line above:

```
$ philote tree --depth 3 shop.philote
shop
  api (backend, hono)
    products: Product
      GET /products
      GET /products/[id]
      POST /products
      PATCH /products/[id]
      DELETE /products/[id]
    customers: Customer
      GET /customers
      GET /customers/[id]
      POST /customers
    orders: Order
      GET /orders
      GET /orders/[id]
  data (3 entities)
    Product
    Customer
    Order
```

`philote expand` shows any part one level down. Creating a product takes a name of at most 120 characters and a price and stock of at least 0. The slug is made from the name:

```
$ philote expand shop.philote api/products
endpoint products.create {
  route: POST "/products"
  input: { name: string(max 120), price_cents: int(min 0), stock: int(min 0) }
  output: Product
  errors: invalid_input
}
```

Philote doesn't know what a shop is. Everything it builds comes from the lines above, so building again gives the same files byte for byte, and changing one line only changes the code for that line.

## Try it

```sh
go install github.com/3nd3r1/philote/cmd/philote@latest
philote check shop.philote    # errors point at the spec line
philote build shop.philote    # writes the server to build/api
cd build/api && npm install && npm start
```

To work on philote, see [CONTRIBUTING.md](CONTRIBUTING.md).
