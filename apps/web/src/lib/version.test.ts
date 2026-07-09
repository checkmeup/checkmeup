import { describe, expect, it } from 'vitest'
import { shortVersion } from './version'

describe('shortVersion', () => {
  it('strips the commit-count/hash suffix from a git describe string', () => {
    expect(shortVersion('v1.22.2-11-g6feb70e')).toBe('v1.22.2')
  })

  it('strips a -dirty suffix too', () => {
    expect(shortVersion('v1.22.2-11-g6feb70e-dirty')).toBe('v1.22.2')
  })

  it('leaves an exact tag (no suffix) unchanged', () => {
    expect(shortVersion('v1.22.2')).toBe('v1.22.2')
  })

  it('falls back to the raw string when it does not look like a version tag', () => {
    expect(shortVersion('dev')).toBe('dev')
  })
})
