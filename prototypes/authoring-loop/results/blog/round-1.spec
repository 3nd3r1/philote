app blog {
  target: nextjs
  style: tokens/minimal
}

entity Post {
  title: string(max 120)
  slug: slug from title
  body: markdown
  published: date
  tags: [Tag]
}

entity Tag {
  label: string unique
}

page "/" {
  data: Post where published <= now order by published desc
  layout: list of PostCard
  empty: "No posts yet."
}

page "/blog/[slug]" {
  data: Post by slug
  layout: PostView
  not_found: 404
}

component PostCard(post: Post) {
  shows: post.title, post.published, excerpt(post.body, 160)
  links_to: "/blog/[post.slug]"
}

component PostView(post: Post) {
  shows: post.title, post.published, post.body as markdown, post.tags as chips
}
