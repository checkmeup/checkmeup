import { post as threeDaysToMvp } from './posts/three-days-to-mvp'
import { post as v1ReleaseChangelog } from './posts/v1-release-changelog'

export type { ContentBlock, BlogPost } from './types'

export const posts = [threeDaysToMvp, v1ReleaseChangelog]

export function getPost(slug: string) {
  return posts.find((p) => p.slug === slug)
}
