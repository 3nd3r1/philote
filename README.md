<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.png">
  <img src="docs/assets/logo.png" alt="philote" width="380">
</picture>

Can we bring Nix-style determinism to AI-driven coding?

Describe your app as an ontology: the things in it, how they relate, and what you can do with them. Philote compiles it to code, from fixed templates wherever it can. Anything AI writes is pinned in a lockfile, so the same ontology always builds the same code.

Status: early. It builds a Hono backend. AI-written steps and a Next.js frontend are next.

## Example

Tinder, for horses. You write what the app is: its data, how the data relates, and which parts the API exposes.

```
app neighbr {
  backend api: hono {
    resource horses: Horse
    resource swipes: Swipe { only: list, create }
    resource matches: Match { only: list, get }
  }
}

enum Breed { arabian, mustang, shetland, clydesdale, other }

entity Horse {
  name: string(max 60)
  handle: slug from name
  breed: Breed
  age: int(min 1, max 40)
  bio: string(max 280)
}

entity Swipe {
  liked: bool
}

entity Match {
  matched_at: datetime
}

Horse has many swipes: Swipe
Swipe belongs to target: Horse (back swiped_by)
Match belongs to first: Horse (back first_matches)
Match belongs to second: Horse (back second_matches)
```

`philote build neighbr.philote` turns that into a runnable Hono server. Every route, rule and error comes from a line above:

```
$ philote tree --depth 3 neighbr.philote
neighbr
  api (backend, hono)
    horses: Horse
      GET /horses
      GET /horses/[id]
      POST /horses
      PATCH /horses/[id]
      DELETE /horses/[id]
    swipes: Swipe
      GET /swipes
      POST /swipes
    matches: Match
      GET /matches
      GET /matches/[id]
  data (3 entities)
    Horse
    Swipe
    Match
```

`philote expand` shows any part one level down. A new horse needs a name of at most 60 characters, a known breed and an age from 1 to 40. The handle is made from the name:

```
$ philote expand neighbr.philote api/horses
endpoint horses.create {
  route: POST "/horses"
  input: { name: string(max 60), breed: Breed, age: int(min 1, max 40), bio: string(max 280) }
  output: Horse
  errors: invalid_input
}
```

Philote doesn't know what a horse is, or what Tinder is. Everything it builds comes from the lines above, so building again gives the same files byte for byte, and changing one line only changes the code for that line.

## Try it

```sh
go install github.com/3nd3r1/philote/cmd/philote@latest
philote check neighbr.philote  # errors point at the spec line
philote build neighbr.philote  # writes the server to build/api
cd build/api && npm install && npm start
```

To work on philote, see [CONTRIBUTING.md](CONTRIBUTING.md).
