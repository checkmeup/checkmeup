import { useHead, useSeoMeta } from '@unhead/vue'
import type { MaybeRefOrGetter } from 'vue'
import { toValue } from 'vue'

const SITE_URL = 'https://checkmeup.net'
const DEFAULT_OG_IMAGE = `${SITE_URL}/img/checkmeup-og.png`

interface SeoOptions {
  title: MaybeRefOrGetter<string>
  description: MaybeRefOrGetter<string>
  path: MaybeRefOrGetter<string>
  image?: MaybeRefOrGetter<string>
}

// One URL form per page: no query string or hash, no trailing slash (except
// the root). The server 301s trailing-slash URLs to this form (handleSPA in
// apps/api/internal/server/server.go), and the sitemap and prerendered pages
// use it too. Echoing the requested URL instead (as route.fullPath did) let
// /blog/x/ and /blog/x?utm_source=… each declare themselves canonical, i.e.
// separate duplicate pages to a search engine.
export function canonicalUrl(path: string): string {
  const bare = path.split(/[?#]/)[0]
  const trimmed = bare.replace(/\/+$/, '')
  return `${SITE_URL}${trimmed || '/'}`
}

// Centralizes the per-page <title>/description/OG/canonical tags every
// marketing and blog view needs — @unhead/vue only updates the live DOM
// (no SSR here), so this fixes Google indexing and browser tab titles but
// not social-card unfurls for links shared before a crawl (see SEO thread).
export function useSeo({ title, description, path, image }: SeoOptions) {
  const url = () => canonicalUrl(toValue(path))
  const ogImage = () => toValue(image) ?? DEFAULT_OG_IMAGE

  useSeoMeta({
    title,
    description,
    ogTitle: title,
    ogDescription: description,
    ogUrl: url,
    ogImage,
    twitterCard: 'summary_large_image',
    twitterTitle: title,
    twitterDescription: description,
    twitterImage: ogImage,
  })

  useHead({
    link: [{ rel: 'canonical', href: url }],
  })
}
