import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "A cron monitor has always had exactly one signal: ping the URL when the job finishes. That catches a job that never runs at all — it misses its window, the grace period expires, you get an alert. It doesn't catch a job that starts, gets stuck in a loop, and just never finishes. No missed ping, because the job never stops to miss one; it's still \"on schedule\" as far as checkmeup can tell, quietly burning CPU or piling up a lock nobody else can take. v1.41 closes that gap: a job can now tell checkmeup when it starts, not just when it's done.",
  },
  {
    type: 'h3',
    text: 'A start ping, entirely optional',
  },
  {
    type: 'p',
    text: 'Every cron monitor gets a second URL — the same ping token, with /start on the end. Call it when your job begins, keep calling the existing URL when it finishes, and that\'s it. Nothing else changes: a job that never touches the new URL behaves exactly as it always has, single ping, no concept of "in progress." This was the one hard requirement going in — five years of monitors already work the old way, and none of them should notice this shipped.',
  },
  {
    type: 'h3',
    text: 'Set how long a run should take',
  },
  {
    type: 'p',
    text: "Once a monitor is sending start pings, its edit form gets a new field: max run duration. Leave it off (the default) and nothing changes. Set it — say, 30 minutes for a job that normally takes five — and checkmeup's existing 30-second worker tick starts checking in-progress runs against that budget alongside its usual missed-ping sweep. Cross it, and you get an alert that says so explicitly: a stuck run, not a missed ping, with when it started and how long it's been going. Recover — either the job finishes, or it starts again fresh and supersedes the old attempt — and a separate recovery alert clears it.",
  },
  {
    type: 'h3',
    text: 'Duration shows up in the execution log',
  },
  {
    type: 'p',
    text: "For any run that sent both pings, the execution log now shows how long it actually took — start ping to completion ping, right next to the existing timestamp and source IP. That's the number you need to pick a sensible max duration in the first place, rather than guessing at one before you've seen how the job actually behaves. A completion ping with no matching start ping just shows a dash, same as any other missing value in this app.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'The start-ping signal and its underlying run history are a new, dedicated table — not a repurposed column on the existing monitor row — so a future release can build overlap detection (catching a job that starts again before the last run finished) on the exact same data without another migration.',
      'Old run records get pruned after 30 days, the same retention window this app already uses for ping history.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Overlap detection is next up, building on the same start-ping signal this release introduces. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
