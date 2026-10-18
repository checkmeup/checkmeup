// @vitest-environment node
// public/robots.txt is served as is, so it's checked straight from disk.
import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { releasesMeta } from '@/blog/posts'

const robots = readFileSync(new URL('../public/robots.txt', import.meta.url), 'utf-8')

// Each Disallow rule as a regex: a path prefix where `*` matches anything
// and a trailing `$` anchors the end, as crawlers read it.
const disallowRules = robots
  .split('\n')
  .map((line) => /^\s*disallow:\s*(\S*)/i.exec(line)?.[1] ?? '')
  .filter((rule) => rule !== '')
  .map((rule) => new RegExp('^' + rule.replace(/[.+?^{}()|[\]\\]/g, '\\$&').replace(/\*/g, '.*')))

describe('robots.txt', () => {
  it("doesn't disallow release notes", () => {
    expect(releasesMeta.length).toBeGreaterThan(0)
    for (const release of releasesMeta) {
      const path = `/blog/${release.slug}`
      const blocking = disallowRules.filter((rule) => rule.test(path))
      expect(blocking, path).toEqual([])
    }
  })
})
