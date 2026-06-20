import { post as threeDaysToMvp } from './posts/three-days-to-mvp'
import { post as v1ReleaseChangelog } from './posts/v1-release-changelog'
import { post as v11ReleaseNotes } from './posts/v1-1-release-notes'
import { post as v12ReleaseNotes } from './posts/v1-2-release-notes'
import { post as pomodoroAndCheckmeup } from './posts/pomodoro-and-checkmeup'
import { post as v13ReleaseNotes } from './posts/v1-3-release-notes'
import { post as v14ReleaseNotes } from './posts/v1-4-release-notes'
import { post as v15ReleaseNotes } from './posts/v1-5-release-notes'

export type { ContentBlock, BlogPost } from './types'

export const posts = [
  threeDaysToMvp,
  v1ReleaseChangelog,
  v11ReleaseNotes,
  v12ReleaseNotes,
  pomodoroAndCheckmeup,
  v13ReleaseNotes,
  v14ReleaseNotes,
  v15ReleaseNotes,
]

export function getPost(slug: string) {
  return posts.find((p) => p.slug === slug)
}
