import type { BlogPost, ContentBlock } from './types'
import { postsMeta as unsortedPostsMeta } from './postsMeta'

export type { ContentBlock, BlogPost, BlogPostMeta } from './types'

// postsMeta.ts's array order is whatever order its entries happen to be
// written in (not meaningful — nothing enforces it), so every consumer needs
// a real chronological sort rather than assuming array order already is one.
export const postsMeta = [...unsortedPostsMeta].sort(
  (a, b) => Date.parse(a.date) - Date.parse(b.date),
)

// Nothing eagerly imports from posts/*.ts (see postsMeta.ts's comment for
// why that matters) — each post's full content is a genuine, separate,
// lazily-loaded chunk, only fetched when a reader actually opens that post.
// Typed Partial<Record<...>> (rather than the plain Record import.meta.glob
// infers) so indexing by a path that doesn't exist — a postsMeta `file`
// typo'd out of sync with the real posts/*.ts filenames, exactly the class
// of bug this codebase has already hit once (see the sitemap
// blog-slug-mismatch fix) — type-checks as possibly undefined instead of
// silently assumed to always succeed (this project doesn't enable
// noUncheckedIndexedAccess, which would otherwise catch this project-wide).
const postLoaders = import.meta.glob<{ content: ContentBlock[] }>('./posts/*.ts') as Partial<
  Record<string, () => Promise<{ content: ContentBlock[] }>>
>

export async function getPost(slug: string): Promise<BlogPost | undefined> {
  const entry = postsMeta.find((m) => m.slug === slug)
  if (!entry) return undefined
  const loader = postLoaders[`./posts/${entry.file}.ts`]
  // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition -- false
  // positive: `tsc --strict` on this exact expression genuinely errors
  // ("Cannot invoke an object which is possibly 'undefined'") without this
  // guard, confirming Partial<Record<...>> above does make `loader` real,
  // necessary `X | undefined` — Codacy's isolated ESLint environment isn't
  // resolving that cast the same way the project's own tsc does (same class
  // of type-resolution gap as no-redundant-type-constituents, see CLAUDE.md).
  if (!loader) return undefined
  const mod = await loader()
  return { ...entry, content: mod.content }
}
