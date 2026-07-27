---
name: release-notes
description: Scaffold a new release-notes blog post under apps/web/src/blog/posts when a version ships, following this repo's existing vX.Y post pattern (a content-only ContentBlock[] file, registered as an entry in postsMeta.ts). Use when asked to "write release notes for vX.Y", "add a blog post for this release", or "ship a blog post for this version".
---

# Release notes blog post

Every shipped version gets a post in `apps/web/src/blog/posts/`, in the
voice of the founder ("Andrew"), registered in
`apps/web/src/blog/postsMeta.ts`. See existing posts for tone — e.g.
`v1-19-release-notes.ts`.

Metadata and content are deliberately split across two files — see
`postsMeta.ts`'s own header comment. Content is lazy-loaded per post via
`import.meta.glob` in `posts.ts` so Rollup code-splits each post into its
own chunk; a post file that also got eagerly imported anywhere (even just
one named export) would force its whole module, content included, into
the eager bundle instead. Don't add an `import` of a post file into
`posts.ts` — that reintroduces exactly that bug.

## Steps

**1. Figure out what shipped.** Use recent commits and the current month's
`docs/reports/YYYY-MM.md` "Shipped" section — that section is written in the
same explanatory voice this post needs, so it's the fastest source of truth
for *what* changed and *why*.

```bash
git log --oneline <since-last-release-tag-or-commit>..HEAD
```

**2. Create the content file** at
`apps/web/src/blog/posts/v<major>-<minor>-release-notes.ts` (dots become
dashes: `v1.19` → `v1-19-release-notes.ts`; a version range like v1.20–v1.22
becomes `v1-20-to-v1-22-redesign.ts` — pick a descriptive suffix if the
release has a theme, otherwise `-release-notes`). It exports **only**
`content` (from `apps/web/src/blog/types.ts`'s `ContentBlock[]`) — no
`slug`/`title`/`date`/etc. here, those live in `postsMeta.ts`:

```ts
import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  { type: 'p', text: '...' },
  { type: 'h3', text: '...' },
  // 'h2', 'ul' (items: string[]), 'code' (lang, text), 'table' (headers, rows),
  // 'blockquote', 'divider', 'image' (src, alt, caption?) also available
  { type: 'h2', text: 'Also this release' },
  { type: 'ul', items: ['minor thing 1', 'minor thing 2'] },
  { type: 'h2', text: 'Follow along' },
  { type: 'p', text: 'What\'s next + pointer to GitHub for full history/ADRs.' },
  { type: 'signature', text: '— Andrew' },
]
```

**3. Register the post** by adding one entry to the `postsMeta` array in
`apps/web/src/blog/postsMeta.ts` (array order doesn't matter — `posts.ts`
sorts by `date` — so add it anywhere, e.g. next to the most recent entry):

```ts
{
  file: 'v1-XX-release-notes',
  slug: 'v1-XX-release-notes',
  title: 'vX.Y: <short thing that shipped>',
  date: 'Month D, YYYY',
  readTime: '<N> min read',
  excerpt: '<2-3 sentence summary, written for someone who has never seen the product>',
},
```

`file` must match the content file's basename under `posts/` exactly —
it's how `posts.ts`'s loader map finds it at read time; a typo here fails
silently as a missing post rather than a type error.

Keep `title` at **52 characters or fewer**. `BlogPostView.vue` renders the
`<title>` tag as `` `${title} — Checkmeup blog` `` (18 more characters), and
SEO checkers flag anything over ~70 total as truncated in search results —
several existing posts (e.g. the v1.29, v1.1, and "Checkmeup vs. ..." posts)
already exceed this and are on the list to shorten.

**4. Voice, from reading existing posts:**

- First person, present tense, conversational — explains *why* a design
  choice was made, not just *what* shipped (e.g. "The answer is the same
  one checkmeup already uses for monitors and status pages: a flat quota
  bundled into the plan price, not metered pass-through billing.")
- Calls out what was deliberately **not** built and why (a scoped-down
  feature is framed as an honest trade-off, not hidden)
- Ends with a "Follow along" paragraph pointing at what's next on the
  roadmap, then a `signature` block
- A "Also this release" `h2` + `ul` section covers smaller items that don't
  need their own paragraphs

**5. Verify.**

```bash
cd apps/web && bun run test -- BlogView BlogPostView
```

Then start the dev server and visually check the new post renders — see
`docs/reference/design.md` for tokens if any custom styling is needed. Don't
use Playwright for this per this repo's convention; if no run-skill exists
yet, say so explicitly rather than reaching for it.

## Also update per this project's Jul 1 precedent

Shipping a new monitor type or major feature (not every release) also
touched: FAQ, pricing table, homepage hero/feature grid, docs page,
Terms of Service, site footer tagline, `CLAUDE.md`, `README.md`. Only do
this broader pass if the release actually changes user-facing capability,
not for every release-notes post.
