import { post as threeDaysToMvp } from './posts/three-days-to-mvp'
import { post as v1ReleaseChangelog } from './posts/v1-release-changelog'
import { post as v11ReleaseNotes } from './posts/v1-1-release-notes'
import { post as v12ReleaseNotes } from './posts/v1-2-release-notes'
import { post as pomodoroAndCheckmeup } from './posts/pomodoro-and-checkmeup'
import { post as v13ReleaseNotes } from './posts/v1-3-release-notes'
import { post as v14ReleaseNotes } from './posts/v1-4-release-notes'
import { post as v15ReleaseNotes } from './posts/v1-5-release-notes'
import { post as v16ReleaseNotes } from './posts/v1-6-release-notes'
import { post as v17ReleaseNotes } from './posts/v1-7-release-notes'
import { post as v18ReleaseNotes } from './posts/v1-8-release-notes'
import { post as v19ReleaseNotes } from './posts/v1-9-release-notes'
import { post as v110ReleaseNotes } from './posts/v1-10-release-notes'
import { post as v111ReleaseNotes } from './posts/v1-11-release-notes'
import { post as v112ReleaseNotes } from './posts/v1-12-release-notes'
import { post as v113ReleaseNotes } from './posts/v1-13-release-notes'
import { post as v114ReleaseNotes } from './posts/v1-14-release-notes'
import { post as v115ReleaseNotes } from './posts/v1-15-release-notes'
import { post as v116ReleaseNotes } from './posts/v1-16-release-notes'

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
  v16ReleaseNotes,
  v17ReleaseNotes,
  v18ReleaseNotes,
  v19ReleaseNotes,
  v110ReleaseNotes,
  v111ReleaseNotes,
  v112ReleaseNotes,
  v113ReleaseNotes,
  v114ReleaseNotes,
  v115ReleaseNotes,
  v116ReleaseNotes,
]

export function getPost(slug: string) {
  return posts.find((p) => p.slug === slug)
}
