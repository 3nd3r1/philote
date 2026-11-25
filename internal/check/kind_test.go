package check

import (
	"strings"
	"testing"

	"github.com/3nd3r1/philote/internal/diag"
)

const kindsSpec = `app planets {
  backend api: hono {
    resource planets: Planet
    resource travelers: Traveler

    endpoint book {
      route: POST "/book"
      input: { traveler: Traveler, planet: Planet }
      output: Planet
      steps: do_book
    }

    step do_book(traveler: Traveler, planet: Planet) -> booked: Planet {
      reads: Planet
      intent: "book a planet"
    }

    endpoint swipe {
      route: POST "/swipe"
      input: { traveler: Traveler, planet: Planet, liked: bool }
      output: Planet
      steps: do_swipe
    }

    step do_swipe(traveler: Traveler, planet: Planet, liked: bool) -> out: Planet {
      reads: Planet
      intent: "record a swipe"
    }
  }
  frontend web: nextjs {
    uses: api
    theme: editorial
    viewer: Traveler

    page "/" {
      load: planets
      shows: Landing, list of PlanetCard(planets)
      empty: "No planets yet."
    }

    page "/pick" {
      load: travelers
      shows: TravelerPicker(travelers), JoinClub
      empty: "No travelers yet."
    }

    page "/planets/[id]" {
      load: planet from planets by id
      shows: PlanetDetail(planet)
      not_found: "No such planet."
    }

    page "/table" {
      load: planets
      shows: PlanetTable(planets), PlanetRows(planets)
      empty: "No planets yet."
    }

    page "/book" {
      load: planets
      shows: PlanetPicker(planets)
      empty: "No planets yet."
    }

    page "/swipe" {
      load: planets
      shows: PlanetDeck(planets)
      empty: "All seen."
    }

    component Landing: hero {
      says: "Book your next planet.", "Weekend stays across the solar system."
      links_to: "/pick" as "Start exploring"
    }

    component JoinClub: callout {
      says: "New traveler?"
      links_to: "/table" as "Browse planets"
    }

    component PlanetCard(planet: Planet): card {
      shows: planet.name, planet.moons
      links_to: "/planets/[planet.id]"
    }

    component PlanetDetail(planet: Planet): detail {
      shows: planet.name, planet.moons
      links_to: "/book" as "Book"
    }

    component PlanetTable(planets: list of Planet): table {
      shows: planets.name, planets.moons
    }

    component PlanetRows(planets: list of Planet): list {
      shows: planets.name
    }

    component TravelerPicker(travelers: list of Traveler): picker {
      shows: travelers.name
      says: "Who are you?"
      picks: viewer then "/book"
    }

    component PlanetPicker(planets: list of Planet): picker {
      shows: planets.name
      says: "Pick a planet"
      calls: book(traveler: viewer, planet: planets)
    }

    component PlanetDeck(planets: list of Planet): deck {
      shows: planets.name, planets.moons
      says: "All seen.", "Booked!", "Keep browsing"
      calls: swipe(traveler: viewer, planet: planets, liked)
    }
  }
}

entity Planet {
  display: name
  name: string(max 60)
  moons: int(min 0)
}

entity Traveler {
  display: name
  name: string(max 60)
}
`

func TestKindsAndThemesResolve(t *testing.T) {
	app, d := load(map[string]string{"app.philote": kindsSpec})
	if len(d.Items) > 0 {
		t.Fatalf("unexpected diagnostics:\n%s", diag.Render(d.Items, map[string]string{"app.philote": kindsSpec}))
	}
	web := app.Parts[1]
	if web.Theme != "editorial" {
		t.Errorf("theme = %q, want editorial", web.Theme)
	}
	want := map[string]string{
		"Landing":        "hero",
		"JoinClub":       "callout",
		"PlanetCard":     "card",
		"PlanetDetail":   "detail",
		"PlanetTable":    "table",
		"PlanetRows":     "list",
		"TravelerPicker": "picker",
		"PlanetPicker":   "picker",
		"PlanetDeck":     "deck",
	}
	for _, comp := range web.Components {
		if comp.Kind != want[comp.Name] {
			t.Errorf("%s kind = %q, want %q", comp.Name, comp.Kind, want[comp.Name])
		}
	}
}

func TestAllThemesValid(t *testing.T) {
	for _, th := range []string{"plain", "warm", "editorial", "slate"} {
		src := strings.Replace(kindsSpec, "theme: editorial", "theme: "+th, 1)
		app, d := load(map[string]string{"app.philote": src})
		if len(d.Items) > 0 {
			t.Fatalf("theme %s: unexpected diagnostics:\n%s", th, diag.Render(d.Items, map[string]string{"app.philote": src}))
		}
		if app.Parts[1].Theme != th {
			t.Errorf("theme = %q, want %q", app.Parts[1].Theme, th)
		}
	}
}

func TestKindAndThemeErrors(t *testing.T) {
	runErrorCases(t, kindsSpec, []errorCase{
		{"unknown theme", "theme: editorial", "theme: neon", "theme: »neon", "unknown theme 'neon'"},
		{"duplicate theme", "theme: editorial", "theme: editorial\n    theme: warm", "»theme: warm", "already has a theme"},
		{"theme and look", "theme: editorial", "theme: editorial\n    look: \"warm\"", "»look: \"warm\"", "can't have both theme and look"},
		{"unknown kind", "): card {", "): widget {", ": »widget {", "unknown component kind 'widget'"},
		{"look on typed component", "shows: planet.name, planet.moons\n      links_to: \"/planets/[planet.id]\"", "shows: planet.name, planet.moons\n      look: \"fancy\"\n      links_to: \"/planets/[planet.id]\"", "look: »\"fancy\"", "built from a template and doesn't take look"},
		{"card with list", "component PlanetTable(planets: list of Planet): table", "component PlanetTable(planets: list of Planet): card", "PlanetTable(planets: list of Planet): »card", "a card component takes one record"},
		{"table with single", "component PlanetCard(planet: Planet): card", "component PlanetCard(planet: Planet): table", "PlanetCard(planet: Planet): »table", "a table component takes a list"},
		{"hero with record", "component PlanetCard(planet: Planet): card", "component PlanetCard(planet: Planet): hero", "PlanetCard(planet: Planet): »hero", "a hero component takes no record"},
		{"card with says", "shows: planet.name, planet.moons\n      links_to: \"/planets/[planet.id]\"", "shows: planet.name, planet.moons\n      says: \"Hi\"\n      links_to: \"/planets/[planet.id]\"", "says: »\"Hi\"", "a card component doesn't take says"},
		{"detail link without label", `links_to: "/book" as "Book"`, `links_to: "/book"`, `links_to: »"/book"`, "needs a button text"},
		{"table with says", "shows: planets.name, planets.moons", "shows: planets.name, planets.moons\n      says: \"All\"", "says: »\"All\"", "a table component doesn't take says"},
		{"hero too many says", `"Book your next planet.", "Weekend stays across the solar system."`, `"One", "Two", "Three"`, `»"Three"`, "a hero component takes at most 2 says"},
		{"callout too many says", `"New traveler?"`, `"New traveler?", "Extra"`, `»"Extra"`, "a callout component takes at most 1 says"},
		{"picker too many says", `says: "Pick a planet"`, `says: "Pick a planet", "Extra"`, `»"Extra"`, "a picker component takes at most 1 says"},
		{"picker missing picks and calls", "      says: \"Pick a planet\"\n      calls: book(traveler: viewer, planet: planets)\n", "      says: \"Pick a planet\"\n", "PlanetPicker(planets: list of Planet): »picker", "a picker component needs picks or calls"},
		{"picker with extra scalar call", "calls: book(traveler: viewer, planet: planets)", "calls: swipe(traveler: viewer, planet: planets, liked)", "calls: »swipe(", "passes only the tapped Planet record"},
		{"deck missing says", "      says: \"All seen.\", \"Booked!\", \"Keep browsing\"\n", "", "PlanetDeck(planets: list of Planet): »deck", "a deck component needs says"},
		{"deck too many says", `"All seen.", "Booked!", "Keep browsing"`, `"One", "Two", "Three", "Four"`, `»"Four"`, "a deck component takes at most 3 says"},
		{"deck missing calls", "      calls: swipe(traveler: viewer, planet: planets, liked)\n", "", "PlanetDeck(planets: list of Planet): »deck", "a deck component needs calls"},
		{"deck calls without bool", "calls: swipe(traveler: viewer, planet: planets, liked)", "calls: book(traveler: viewer, planet: planets)", "Keep browsing\"\n      calls: »book(", "passes the current Planet record and one bool"},
	})
}
