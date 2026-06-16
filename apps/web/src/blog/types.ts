export type ContentBlock =
  | { type: 'p'; text: string }
  | { type: 'h2'; text: string }
  | { type: 'h3'; text: string }
  | { type: 'code'; lang: string; text: string }
  | { type: 'ul'; items: string[] }
  | { type: 'blockquote'; text: string }
  | { type: 'divider' }
  | { type: 'signature'; text: string }

export interface BlogPost {
  slug: string
  title: string
  date: string
  readTime: string
  excerpt: string
  content: ContentBlock[]
}
