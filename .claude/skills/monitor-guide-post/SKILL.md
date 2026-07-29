---
name: monitor-guide-post
description: Write a new evergreen SEO guide post explaining one monitor type (e.g. "Cron Job Monitoring", "Uptime Monitoring", "Port (TCP) Monitoring") — distinct from release-notes posts, which announce what shipped. Use when asked to "write an SEO post", "write a guide for X monitoring", or "add a monitor-type explainer post"; if no topic was named, this skill's first step is picking one.
---

# Monitor-type guide post

Three of these exist — `cron-job-monitoring-guide`, `uptime-monitoring-guide`,
`port-tcp-monitoring-guide` — sharing one fixed structure and voice, distinct
from `release-notes` posts (which announce a shipped version) and disaster-story
posts like `ssl-certificate-expired-story` (a one-off narrative, not this
pattern). Read one of the three existing guide posts in full before writing —
this file describes the shape, not the prose voice.

## Steps

**1. Pick the topic, if not already named.** Cross-reference CLAUDE.md's
monitor list (cron, uptime, SSL expiry, domain expiry, port/TCP, DNS record) against
existing `*-guide` slugs in `apps/web/src/blog/postsMeta.ts` to find the gap.
If more than one gap exists, ask the user which one — don't guess.

**2. Verify the feature's actual behavior in code before writing a word.**
Marketing copy for a feature that doesn't work the way you assume is worse
than no copy — this project's own guidance (verify feature claims by reading
code, not docs or memory) applies doubly to a post that will sit on the
public site indefinitely. Find the check's implementation (e.g.
`apps/api/internal/worker/worker_<type>.go`), and confirm: what the check
actually does (protocol-aware or transport-only?), what config fields exist,
what counts as failure, what the available check intervals are, and the
current free-tier monitor limit (`docs/reference/limits.md` /
`internal/billing/plans.go`). Note anything the feature does *not* do —
the "what this doesn't prove" section (see step 3) depends on knowing the
real limits, not assumed ones.

**3. Write the content file** at
`apps/web/src/blog/posts/<type>-monitoring-guide.ts`, following this fixed
section order (see the three existing posts for exact tone/length per
section — roughly 900-1100 words total):

1. **Hook paragraph** — a concrete failure scenario this monitor type catches
   that a *different* monitor type wouldn't (e.g. cron: a job that dies
   silently; uptime: a 200 response from a broken page). No heading.
2. **`h2` "What \<type\> monitoring actually does"** — mechanical explanation
   of the check itself, grounded in what you verified in step 2.
3. **`h2` naming the check's real limitation** — what a passing check does
   *not* prove, stated plainly rather than glossed over. This is what makes
   the post trustworthy rather than pure marketing.
4. **`h2` "Setting up \<type\> monitoring in three steps"** — three `h3`
   substeps, each one concrete decision + why it matters, not generic advice.
5. **`h2` "What to look for in a \<type\> monitoring tool"** — a `ul` of
   4-5 vendor-neutral criteria (reads as evaluation advice, not a feature
   list — Checkmeup's own pitch stays out of this section).
6. **`h2` "The mistake worth avoiding"** — one specific, common way people
   get a false sense of security from this monitor type.
7. **`divider`**, then one `p` pitching Checkmeup's actual implementation of
   this monitor type (only claims verified in step 2), ending with the
   Hobby-plan free-tier line other guide posts use.
8. **`signature`**, text `'— Andrew'`.

```ts
import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  { type: 'p', text: '...' },
  { type: 'h2', text: 'What <type> monitoring actually does' },
  { type: 'p', text: '...' },
  // ...
  { type: 'signature', text: '— Andrew' },
]
```

**4. Register in `apps/web/src/blog/postsMeta.ts`** — add one entry to the
`postsMeta` array (order doesn't matter, sorted by date elsewhere):

```ts
{
  file: '<type>-monitoring-guide',
  slug: '<type>-monitoring-guide',
  title: '<Type> Monitoring: <short angle>',   // 52 chars or fewer, see below
  date: 'Month D, YYYY',
  readTime: '5 min read',
  excerpt: '<2-3 sentences: the gap this monitor type fills, and the one caveat about what it doesn\'t prove>',
},
```

`file` must exactly match the content file's basename under `posts/` — the
loader map in `posts.ts` keys on it, and a mismatch fails silently as a
missing post rather than a type error. Keep `title` at **52 characters or
fewer**: `BlogPostView.vue` renders `` `${title} — Checkmeup blog` `` (18
more characters), and SEO checkers flag the combined tag as truncated past
~70.

**5. Verify.**

```bash
cd apps/web && npx oxlint src/blog/posts/<type>-monitoring-guide.ts src/blog/postsMeta.ts
cd apps/web && npx vue-tsc --noEmit -p tsconfig.json
```

## What NOT to claim

Never imply the check does more than step 2 confirmed — e.g. don't say a
port monitor "verifies your database is responding correctly" when it's a
bare TCP connect with no protocol handshake. The limitation section (step 3,
item 3) exists specifically so this doesn't need to be papered over; naming
the gap honestly is what makes the "what to look for" checklist credible.
