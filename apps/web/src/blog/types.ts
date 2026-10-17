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
  // 'release' = a vX.Y release-notes post: stays reachable for readers (via
  // /changelog and its own URL) but is kept out of search — robots
  // noindex in its prerendered HTML, left out of sitemap.xml, and not listed
  // on /blog. Dozens of thin per-version pages were dragging down how search
  // engines rate the whole site; only 'article' posts (guides, essays) are
  // pages people actually search for.
  kind: 'article' | 'release'
  date: string
  readTime: string
  excerpt: string
  // The meta description search results show, 155 characters or fewer —
  // Google cuts it off past ~160, and an excerpt is written for the page,
  // usually twice that. Every 'article' needs one (BlogPostView.test.ts
  // checks); a release post, being noindex, falls back to its excerpt.
  description?: string
}

export interface BlogPost extends BlogPostMeta {
  content: ContentBlock[]
}
