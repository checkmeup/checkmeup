import { post as threeDaysToMvp } from './posts/three-days-to-mvp'
import { post as v10ReleaseNotes } from './posts/v1-0-release-notes'
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
import { post as v117ReleaseNotes } from './posts/v1-17-release-notes'
import { post as v118ReleaseNotes } from './posts/v1-18-release-notes'
import { post as checkmeupVsCompetitors } from './posts/checkmeup-vs-competitors'
import { post as v119ReleaseNotes } from './posts/v1-19-release-notes'
import { post as v120To122Redesign } from './posts/v1-20-to-v1-22-redesign'
import { post as v123ReleaseNotes } from './posts/v1-23-release-notes'
import { post as v124To125SeoFoundations } from './posts/v1-24-to-v1-25-seo-foundations'
import { post as v126ReleaseNotes } from './posts/v1-26-release-notes'
import { post as v127ReleaseNotes } from './posts/v1-27-release-notes'
import { post as v128ReleaseNotes } from './posts/v1-28-release-notes'
import { post as v129ReleaseNotes } from './posts/v1-29-release-notes'

export type { ContentBlock, BlogPost } from './types'

export const posts = [
  threeDaysToMvp,
  v10ReleaseNotes,
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
  v117ReleaseNotes,
  v118ReleaseNotes,
  checkmeupVsCompetitors,
  v119ReleaseNotes,
  v120To122Redesign,
  v123ReleaseNotes,
  v124To125SeoFoundations,
  v126ReleaseNotes,
  v127ReleaseNotes,
  v128ReleaseNotes,
  v129ReleaseNotes,
]

export function getPost(slug: string) {
  return posts.find((p) => p.slug === slug)
}
