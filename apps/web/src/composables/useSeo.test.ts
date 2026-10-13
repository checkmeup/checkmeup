import { describe, expect, it } from 'vitest'
import { canonicalUrl } from './useSeo'

describe('canonicalUrl', () => {
  it('keeps the root as a single slash', () => {
    expect(canonicalUrl('/')).toBe('https://checkmeup.net/')
  })

  it('leaves an already-canonical path unchanged', () => {
    expect(canonicalUrl('/blog/cron-job-monitoring-guide')).toBe(
      'https://checkmeup.net/blog/cron-job-monitoring-guide',
    )
  })

  it('drops a trailing slash so both URL forms share one canonical', () => {
    expect(canonicalUrl('/blog/v1-28-release-notes/')).toBe(
      'https://checkmeup.net/blog/v1-28-release-notes',
    )
    expect(canonicalUrl('/blog//')).toBe('https://checkmeup.net/blog')
  })

  it('drops the query string and hash', () => {
    expect(canonicalUrl('/blog/x?utm_source=newsletter')).toBe('https://checkmeup.net/blog/x')
    expect(canonicalUrl('/changelog/#v1-41-zombie-job-detection')).toBe(
      'https://checkmeup.net/changelog',
    )
  })
})
