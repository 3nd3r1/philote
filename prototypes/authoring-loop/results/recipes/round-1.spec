// gap: search (no way to search recipes by title)
// gap: per-category pages listing their recipes (list pages can't take route params, and a list of recipes can't be shown on a single Category page)

app recipes {
  target: nextjs
  style: tokens/minimal
}

entity Recipe {
  title: string(max 120)
  slug: slug from title
  description: string(max 280)
  ingredients: markdown
  instructions: markdown
  cook_minutes: int
  category: Category
}

entity Category {
  name: string unique
}

page "/" {
  data: Recipe order by title asc
  layout: list of RecipeCard
  empty: "No recipes yet."
}

page "/recipes/[slug]" {
  data: Recipe by slug
  layout: RecipeDetail
  not_found: 404
}

component RecipeCard(recipe: Recipe) {
  shows: recipe.title, recipe.description, recipe.cook_minutes, recipe.category.name
  links_to: "/recipes/[recipe.slug]"
}

component RecipeDetail(recipe: Recipe) {
  shows: recipe.title, recipe.description, recipe.category.name, recipe.cook_minutes, recipe.ingredients as markdown, recipe.instructions as markdown
}
