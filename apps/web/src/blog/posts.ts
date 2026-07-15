import type { BlogPost, ContentBlock } from './types'
import { postsMeta } from './postsMeta'

export type { ContentBlock, BlogPost, BlogPostMeta } from './types'
export { postsMeta } from './postsMeta'

// Nothing eagerly imports from posts/*.ts (see postsMeta.ts's comment for
// why that matters) — each post's full content is a genuine, separate,
// lazily-loaded chunk, only fetched when a reader actually opens that post.
const postLoaders = import.meta.glob<{ content: ContentBlock[] }>('./posts/*.ts')

export async function getPost(slug: string): Promise<BlogPost | undefined> {
  const entry = postsMeta.find((m) => m.slug === slug)
  if (!entry) return undefined
  const loader = postLoaders[`./posts/${entry.file}.ts`]
  if (!loader) return undefined
  const mod = await loader()
  return { ...entry, content: mod.content }
}
