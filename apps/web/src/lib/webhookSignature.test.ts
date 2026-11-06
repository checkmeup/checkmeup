import { createRequire } from 'node:module'
import { describe, expect, it } from 'vitest'
import { signatureSnippet } from './webhookSignature'

// Computed by the API's webhook.Sign for this body and secret.
const body = '{"eventType":"down","monitorName":"api.example.com"}'
const secret = 'shh-secret'
const signature = 'f1366d9fcba07d97535f34e6a5ac072da259cf7d5fef571cee7713686ace6ac0'

const isFromCheckmeup = new Function('require', `${signatureSnippet}\nreturn isFromCheckmeup`)(
  createRequire(import.meta.url),
) as (rawBody: string, signature: unknown, secret: string) => boolean

describe('signatureSnippet', () => {
  it('accepts a request the API signed', () => {
    expect(isFromCheckmeup(body, signature, secret)).toBe(true)
  })

  it('rejects a changed body, a wrong secret, and a missing or malformed header', () => {
    expect(isFromCheckmeup(body.replace('down', 'recovery'), signature, secret)).toBe(false)
    expect(isFromCheckmeup(body, signature, 'other-secret')).toBe(false)
    expect(isFromCheckmeup(body, undefined, secret)).toBe(false)
    expect(isFromCheckmeup(body, 'abc', secret)).toBe(false)
  })
})
