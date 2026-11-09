// gap: no forms or actions, so readers can't leave a comment on a post
// gap: no way to show a post's comments under it, oldest first (no related-record queries inside a single-record page)
// gap: no way to list an author's posts on their author page (no reverse relation from Author to Post)

app blog {
  target: nextjs
  style: tokens/minimal
}

entity Author {
  name: string(max 80)
  handle: slug from name
  bio: markdown
}

entity Post {
  title: string(max 120)
  slug: slug from title
  body: markdown
  published: date
  author: Author
}

page "/posts" {
  data: Post where published <= now order by published desc
  layout: list of PostCard
  empty: "No posts yet."
}

page "/posts/[slug]" {
  data: Post by slug
  layout: PostArticle
  not_found: 404
}

page "/authors/[handle]" {
  data: Author by handle
  layout: AuthorProfile
  not_found: 404
}

component PostCard(post: Post) {
  shows: post.title, post.author.name, post.published, excerpt(post.body, 160)
  links_to: "/posts/[post.slug]"
}

component PostArticle(post: Post) {
  shows: post.title, post.author.name, post.published, post.body as markdown
}

component AuthorProfile(author: Author) {
  shows: author.name, author.bio as markdown
}
