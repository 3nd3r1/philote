app team {
  target: nextjs
  style: tokens/minimal
}

entity Person {
  name: string(max 80)
  handle: slug from name
  bio: markdown
  joined: date
  skills: [Skill]
}

entity Skill {
  label: string unique
}

page "/people" {
  data: Person order by name asc
  layout: list of PersonCard
  empty: "Nobody here yet."
}

page "/people/[handle]" {
  data: Person by handle
  layout: PersonProfile
  not_found: 404
}

component PersonCard(person: Person) {
  shows: person.name, excerpt(person.bio, 120)
  links_to: "/people/[person.handle]"
}

component PersonProfile(person: Person) {
  shows: person.name, person.joined, person.bio as markdown, person.skills as chips
}
