| Brief | Rounds | Clean | Cost | Time | Gaps reported |
| --- | --- | --- | --- | --- | --- |
| blog-authors-comments | 1 | true | $0.048 | 17s | no forms or actions, so readers can't leave a comment on a post; no way to show a post's comments under it, oldest first (no related-record queries inside a single-record page); no way to list an author's posts on their author page (no reverse relation from Author to Post) |
| blog | 1 | true | $0.024 | 5s |  |
| recipes | 1 | true | $0.039 | 13s | search (no way to search recipes by title); per-category pages listing their recipes (list pages can't take route params, and a list of recipes can't be shown on a single Category page) |
