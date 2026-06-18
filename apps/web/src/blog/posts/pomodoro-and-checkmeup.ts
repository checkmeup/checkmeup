import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'pomodoro-and-checkmeup',
  title: 'The Cube on My Desk: Pomodoro and Building checkmeup',
  date: 'June 18, 2026',
  readTime: '3 min read',
  excerpt:
    "A small black cube sits next to my keyboard and ticks down in 25-minute blocks. Here's what the Pomodoro Technique is, and how it shaped the pace checkmeup got built at.",
  content: [
    {
      type: 'p',
      text: 'This cube sits next to my keyboard. It has four faces — 5, 10, 25, and 50 minutes. Flip it to whichever face you want and it counts down. No app, no phone, no notifications fighting for attention. Just a number going down.',
    },
    {
      type: 'image',
      src: '/blog/pomodoro-and-checkmeup.jpg',
      alt: 'A black cube-shaped Pomodoro timer showing 21:33 on its display, resting in a hand next to a mechanical keyboard and a laptop.',
      caption: '25 minutes of focus, then a break before the next one.',
    },
    {
      type: 'h2',
      text: 'What the Pomodoro Technique actually is',
    },
    {
      type: 'p',
      text: 'Francesco Cirillo came up with it in the late 1980s using a kitchen timer shaped like a tomato — pomodoro, in Italian. The mechanics are almost embarrassingly simple: pick one task, set a timer for 25 minutes, work on only that task until it rings. Then take a 5-minute break. After four of those cycles, take a longer one — 15 to 30 minutes.',
    },
    {
      type: 'p',
      text: "That's it. The value isn't the timer itself — it's the forced single-tasking. For 25 minutes, switching to email, checking a notification, or going down a refactoring rabbit hole that wasn't the task at hand simply isn't allowed. The clock is the commitment device.",
    },
    {
      type: 'h2',
      text: 'How it shaped checkmeup',
    },
    {
      type: 'p',
      text: "Checkmeup wasn't built in long, uninterrupted weeks. It was built in the margins — roughly 3 to 4 hours on weekdays, a bit more on weekends, fit around everything else life requires. With that little uninterrupted time per sitting, every Pomodoro has to count.",
    },
    {
      type: 'p',
      text: "The discipline showed up in an unexpected place: the project's own hour log, which says the MVP took 29 hours. Every one of those hours went into deciding exactly what to build, reviewing what came back, and redirecting it when the AI confidently shipped the wrong thing at impressive speed. That's the part Pomodoro actually trained — not typing faster, but deciding, before the timer starts, what the one task for this block is. Hand an AI a vague goal and it will cheerfully build a large, vague, occasionally hilarious solution to a problem you didn't quite have.",
    },
    {
      type: 'p',
      text: "One focused block is enough to decide exactly what a migration and its query need to do — the AI writes them in about the time it takes to read this sentence twice. Two blocks is enough to scope a handler and its route the same way. When a feature eats more than four or five blocks of my attention in a single sitting, that's not the AI being slow — it's a sign I handed it something too big and too vague to aim properly. Which, not coincidentally, is exactly why this project's epics and user stories ended up broken into pieces that small in the first place.",
    },
    {
      type: 'blockquote',
      text: "The timer doesn't write the code. It just makes sure the 25 minutes you do have aren't wasted deciding what to work on.",
    },
    {
      type: 'h2',
      text: 'Why a physical cube and not an app',
    },
    {
      type: 'p',
      text: "A phone timer means a phone on the desk, and a phone on the desk means notifications eventually winning. The cube does one thing. Flip it, it counts down, it has no Wi-Fi. That's the whole feature set, and it's the right one.",
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
