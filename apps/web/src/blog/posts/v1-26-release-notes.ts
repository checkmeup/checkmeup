import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: 'Three unrelated-sounding things shipped this week, but they share a theme: finishing things properly instead of leaving the edges rough. A visitor hitting a dead link deserves a real page, not a blank one. A blog post shared as a link deserves to look like it belongs to Checkmeup when Google shows it in search results. And "Checkmeup" deserves to actually read as a proper noun, not drift between capitalized and lowercase depending on which file you\'re looking at.',
  },
  {
    type: 'h3',
    text: 'A 404 page that was actually missing',
  },
  {
    type: 'p',
    text: "Until this release, only one route in the entire app had a fallback for a bad URL: an unknown blog slug. Everything else — a typo'd path, an old bookmark, a link that used to work — had nothing. Vue Router had no catch-all route, so hitting anything unmatched just rendered a blank page. It's not a huge deal traffic-wise, but it's the kind of gap that looks bad the one time someone actually hits it, and it was silently missing an SEO fix too: a page that returns HTTP 200 with nothing on it is exactly the kind of thin content a search engine shouldn't index.",
  },
  {
    type: 'p',
    text: 'The fix adds a real catch-all route and a proper 404 page — implemented from an actual Claude Design mockup rather than a placeholder, with the grid-texture background and accent glow the homepage\'s hero section already uses, so it doesn\'t look like it was bolted on. It sets a noindex tag unconditionally, same mechanism as the blog-slug fix already had. And since the visual design turned out to be worth reusing, the blog\'s own "post not found" state got upgraded to match instead of staying as a plain line of text — same component, different copy: "Back to dashboard" for a dead link anywhere in the app, "Back to blog" for a bad post slug.',
  },
  {
    type: 'h3',
    text: 'Closing out the SEO thread: Article schema',
  },
  {
    type: 'p',
    text: "v1.24–v1.25 added FAQPage structured data so the FAQ page can show up as an expandable result in Google search. The same idea now applies to every blog post: Article JSON-LD with the post's headline, description, author, and publish date, which is what makes a shared post eligible for a richer search listing instead of a plain blue link. It only appears for a post that actually exists — an unknown slug gets nothing, consistent with that state already being noindexed.",
  },
  {
    type: 'p',
    text: "Worth being honest about what this doesn't fix: none of it touches how a link looks when it's pasted into Slack or Twitter. Those previews come from bots that don't run JavaScript, so they still see the site's generic title and image no matter which post got shared — Article schema and the per-page meta tags from v1.24 only help search engines, which do execute JS. Fixing social previews needs the page's raw HTML to actually differ per post, which means prerendering at build time — real architectural work, filed as its own backlog item (EP-36) rather than rushed into this release.",
  },
  {
    type: 'h3',
    text: "Getting Checkmeup's own name right",
  },
  {
    type: 'p',
    text: 'This one\'s less exciting and more overdue: "checkmeup" and "Checkmeup" had been used interchangeably in prose across the entire codebase since launch — sometimes capitalized, sometimes not, with no rule anyone had written down. Email subject lines, Telegram and SMS alert text, blog posts, marketing copy — all inconsistent. The convention is now explicit and documented in CLAUDE.md: capitalize it as the organization\'s name in prose, keep the checkmeup.net domain lowercase mid-sentence (capitalizing only when it opens a sentence), and never touch the lowercase form where it\'s actually a code identifier — the GitHub org slug, the npm package scope, a sessionStorage key, an asset filename. None of those are prose; renaming them would\'ve been a bug, not a fix.',
  },
  {
    type: 'p',
    text: 'Applying it touched a genuinely large surface — every blog post, the marketing and docs pages, and the backend strings that actually reach a person: email subjects and bodies, Telegram and Slack alert text, SMS copy, the public status page\'s "Powered by" link. Deliberately left alone: a couple of Go doc comments that are developer-facing, not something a user ever sees.',
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'The very first release post — originally filed as "v1-release-changelog" — got renamed to v1-0-release-notes to match every later version\'s filename/slug pattern. Its content didn\'t change, only where it lives.',
      'Found and fixed a bug in the capitalization pass itself while double-checking it: the exclusion rule protecting the checkmeup.net domain was too broad and also skipped plain "checkmeup." at the end of an unrelated sentence — caught before it shipped.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "Next up is a shift from technical SEO plumbing to actual content — the blog is still almost entirely release notes, which serves existing users well but does nothing for someone searching for a monitoring tool who's never heard of Checkmeup. Full commit history and the architecture decisions behind Checkmeup are public on GitHub if you want the details behind any of this.",
  },
  { type: 'signature', text: '— Andrew' },
]
