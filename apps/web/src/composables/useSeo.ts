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

// Centralizes the per-page <title>/description/OG/canonical tags every
// marketing and blog view needs — @unhead/vue only updates the live DOM
// (no SSR here), so this fixes Google indexing and browser tab titles but
// not social-card unfurls for links shared before a crawl (see SEO thread).
export function useSeo({ title, description, path, image }: SeoOptions) {
  const url = () => `${SITE_URL}${toValue(path)}`
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
