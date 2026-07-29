import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "The public status page has had exactly one look since it shipped: a single column, monitors stacked top to bottom, incidents inline above them. That's still the right default for most pages. But a page watching a dozen monitors turns into a long scroll, and the incident history ends up buried at the bottom of it. v1.40 adds a second layout built for that case, and lets you pick per page which one fits.",
  },
  {
    type: 'h3',
    text: 'Classic and grid',
  },
  {
    type: 'p',
    text: '"Classic" is the layout that\'s always been there, unchanged — every existing status page keeps it by default. "Grid" lays monitors out two-per-row in a wider page, and moves active incidents and past-incident history into a sidebar next to them instead of above and below. It reads better once you\'re past four or five monitors, where the classic layout\'s vertical scroll starts working against you.',
  },
  {
    type: 'p',
    text: 'Layout is a setting on the page, not the org — the same status page edit form where you already control the title, description, logo, and branding. Run more than one status page and each can make its own call: a client-facing page might stay classic for simplicity, while an internal one tracking everything gets the grid.',
  },
  {
    type: 'h3',
    text: 'Still server-rendered, still no JS required',
  },
  {
    type: 'p',
    text: "The status page has rendered as plain server-side HTML since it launched, specifically so it works for visitors, curl, and status-checking bots with JavaScript disabled or absent entirely. That didn't change to add a second layout — both are the same Go template, branching on which one to render, so grid gets the same no-JS guarantee classic always had.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      "Layout isn't gated to any plan — it's a display preference, not a paid feature, so it's available whether you're on Hobby or Enterprise.",
      "Fixed a TypeScript type gap in the status page API client that hadn't been updated when DNS monitors shipped — no runtime bug, but it would have blocked a clean type-check on the next status page change that touched DNS monitors.",
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "That's two layouts now, with room for more if a third shape turns out to earn its keep. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
