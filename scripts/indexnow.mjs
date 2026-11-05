// Tells Bing, Yandex and the other IndexNow engines which public pages a deploy
// changed (ADR seo). The release workflow runs it twice around `kamal deploy`:
//
//   node scripts/indexnow.mjs snapshot > before.json   # before the deploy
//   node scripts/indexnow.mjs submit before.json       # after it
//
// A snapshot hashes the live HTML of every URL in the live sitemap.xml, with the
// build's /assets/ file names and version numbers (the footer's) taken out, since
// they change on every build. Submit takes a new snapshot and sends the URLs whose
// hash changed or that are new.
// DRY_RUN=1 prints the URLs instead of sending them.
import { createHash } from 'node:crypto'
import { readFileSync } from 'node:fs'

const SITE = 'https://checkmeup.net'
// Served from apps/web/public/<key>.txt, which proves the site owns the key.
const KEY = 'b76076cc360a44c49982ee5d5143ea9b'

async function snapshot() {
  const sitemap = await (await fetch(`${SITE}/sitemap.xml`)).text()
  const urls = [...sitemap.matchAll(/<loc>([^<]+)<\/loc>/g)].map((m) => m[1])
  const hashes = {}
  for (const url of urls) {
    const res = await fetch(url)
    const html = (await res.text())
      .replace(/\/assets\/[^"']+/g, '/assets/')
      .replace(/\bv\d+\.\d+\.\d+[\w.-]*/g, 'v')
    hashes[url] = `${res.status} ${createHash('sha256').update(html).digest('hex')}`
  }
  return hashes
}

const [command, beforeFile] = process.argv.slice(2)
if (command === 'snapshot') {
  console.log(JSON.stringify(await snapshot()))
} else if (command === 'submit' && beforeFile) {
  const before = JSON.parse(readFileSync(beforeFile, 'utf-8'))
  const after = await snapshot()
  const changed = Object.keys(after).filter((url) => after[url] !== before[url])
  if (changed.length === 0 || process.env.DRY_RUN) {
    console.log(`${changed.length} changed:\n${changed.join('\n')}`)
  } else {
    const res = await fetch('https://api.indexnow.org/indexnow', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json; charset=utf-8' },
      body: JSON.stringify({
        host: new URL(SITE).host,
        key: KEY,
        keyLocation: `${SITE}/${KEY}.txt`,
        urlList: changed,
      }),
    })
    console.log(`IndexNow answered ${res.status} for ${changed.length} URLs:\n${changed.join('\n')}`)
    if (!res.ok) process.exit(1)
  }
} else {
  console.error('usage: node scripts/indexnow.mjs snapshot | submit <before.json>')
  process.exit(2)
}
