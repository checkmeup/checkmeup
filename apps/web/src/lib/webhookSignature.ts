// Shown under a webhook channel's signing secret (PRD webhook-alerts): Node.js
// code that checks X-Checkmeup-Signature, the hex HMAC-SHA256 of the raw body.
// webhookSignature.test.ts runs it against a signature the API computed.
export const signatureSnippet = `const crypto = require('node:crypto')

// rawBody: the request body exactly as received, before JSON parsing
function isFromCheckmeup(rawBody, signature, secret) {
  const expected = crypto.createHmac('sha256', secret).update(rawBody).digest('hex')
  return typeof signature === 'string' && signature.length === expected.length &&
    crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(expected))
}

// isFromCheckmeup(rawBody, req.headers['x-checkmeup-signature'], process.env.CHECKMEUP_WEBHOOK_SECRET)`
