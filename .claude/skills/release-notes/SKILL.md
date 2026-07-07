---
name: release-notes
description: Scaffold a new release-notes blog post under apps/web/src/blog/posts when a version ships, following this repo's existing vX.Y post pattern (BlogPost type, content blocks, registration in posts.ts). Use when asked to "write release notes for vX.Y", "add a blog post for this release", or "ship a blog post for this version".
---

# Release notes blog post

Every shipped version gets a post in `apps/web/src/blog/posts/`, in the
voice of the founder ("Andrew"), registered in `apps/web/src/blog/posts.ts`.
See existing posts for tone — e.g. `v1-19-release-notes.ts`.

## Steps

**1. Figure out what shipped.** Use recent commits and the current month's
`docs/reports/YYYY-MM.md` "Shipped" section — that section is written in the
same explanatory voice this post needs, so it's the fastest source of truth
for *what* changed and *why*.

```bash
git log --oneline <since-last-release-tag-or-commit>..HEAD
```

**2. Create the post file** at
`apps/web/src/blog/posts/v<major>-<minor>-release-notes.ts` (dots become
dashes: `v1.19` → `v1-19-release-notes.ts`; a version range like v1.20–v1.22
becomes `v1-20-to-v1-22-redesign.ts` — pick a descriptive suffix if the
release has a theme, otherwise `-release-notes`).

Shape (from `apps/web/src/blog/types.ts`):

```ts
import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-XX-release-notes',
  title: 'vX.Y: <short thing that shipped>',
  date: 'Month D, YYYY',
  readTime: '<N> min read',
  excerpt: '<2-3 sentence summary, written for someone who has never seen the product>',
  content: [
    { type: 'p', text: '...' },
    { type: 'h3', text: '...' },
    // 'h2', 'ul' (items: string[]), 'code' (lang, text), 'table' (headers, rows),
    // 'blockquote', 'divider', 'image' (src, alt, caption?) also available
    { type: 'h2', text: 'Also this release' },
    { type: 'ul', items: ['minor thing 1', 'minor thing 2'] },
    { type: 'h2', text: 'Follow along' },
    { type: 'p', text: 'What\'s next + pointer to GitHub for full history/ADRs.' },
    { type: 'signature', text: '— Andrew' },
  ],
}
```

**3. Register the post** in `apps/web/src/blog/posts.ts` — two edits, both
appended at the end (chronological order):

```ts
import { post as v1XXReleaseNotes } from './posts/v1-XX-release-notes'
```

and add `v1XXReleaseNotes,` as the last entry in the `posts` array.

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
