import type { BlogPost, ContentBlock } from './types'
import { postsMeta } from './postsMeta'

export type { ContentBlock, BlogPost, BlogPostMeta } from './types'
export { postsMeta } from './postsMeta'

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
  if (!loader) return undefined
  const mod = await loader()
  return { ...entry, content: mod.content }
}
