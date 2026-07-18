export type ContentBlock =
  | { type: 'p'; text: string }
  | { type: 'h2'; text: string }
  | { type: 'h3'; text: string }
  | { type: 'code'; lang: string; text: string }
  | { type: 'ul'; items: string[] }
  | { type: 'blockquote'; text: string }
  | { type: 'divider' }
  | { type: 'signature'; text: string }
  | { type: 'image'; src: string; alt: string; caption?: string }
  | { type: 'table'; headers: string[]; rows: string[][] }

export interface BlogPostMeta {
  slug: string
  // Keep to 52 characters or fewer, for every post (release notes and
  // SEO/content posts alike, whichever workflow wrote it) — BlogPostView.vue
  // renders the <title> tag as `${title} — Checkmeup blog`, 18 more
  // characters, and SEO checkers flag the combined tag as truncated past
  // ~70. Caught late on two posts (cron-job-monitoring-guide,
  // uptime-monitoring-guide) that were drafted without going through the
  // release-notes skill, which already documented this budget for its own
  // output but nothing enforced it here.
  title: string
  date: string
  readTime: string
  excerpt: string
}

export interface BlogPost extends BlogPostMeta {
  content: ContentBlock[]
}
