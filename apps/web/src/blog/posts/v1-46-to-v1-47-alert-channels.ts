import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Two releases, and most of what's in them is about alerts that could fail without anyone finding out. That's the worst kind of bug for a monitoring tool: the whole promise is that you hear about problems, so a problem with the hearing part is the one thing that can't stay quiet.",
  },
  {
    type: 'h2',
    text: 'An SMS that never arrived',
  },
  {
    type: 'p',
    text: "Until now, an alert SMS counted as delivered the moment Twilio accepted it. But accepting a message isn't delivering it: the carrier can still drop it minutes later, and Checkmeup would never know. If SMS was the only channel on that monitor, the alert was simply gone.",
  },
  {
    type: 'p',
    text: "Every alert SMS now asks Twilio to report back its final status. When that status is failed or undelivered, it's recorded on the channel, so Settings shows it. And if the SMS was the alert's only delivery, the fallback email goes out at that point. To make that decision possible, SMS channels are now sent last, after everything else has had its chance.",
  },
  {
    type: 'p',
    text: "One design detail worth explaining: Twilio normally lets you verify its callbacks with a signature, but that signature needs the account's primary Auth Token, and Checkmeup deliberately only holds a scoped API key. So the callback URL carries its own token instead, derived from a server secret. And when Twilio rejects a message outright, the channel now shows Twilio's own reason, such as \"400: The 'To' number … is not a valid phone number\", instead of a bare \"failed, 400\".",
  },
  {
    type: 'h2',
    text: 'Webhook URLs are secrets',
  },
  {
    type: 'p',
    text: "A Slack incoming-webhook URL, or your own webhook endpoint, is effectively a password: anyone who has it can post to that channel. The API used to send those URLs back in full every time the settings page loaded. Now it returns them masked, like https://hooks.slack.com/…abcd. Editing a channel and leaving the URL field empty keeps the stored one, so renaming a channel doesn't mean pasting the secret again.",
  },
  {
    type: 'h2',
    text: 'A missing page now says so',
  },
  {
    type: 'p',
    text: "Any path the web app had no route for, say /nope, used to get the homepage HTML with a 200 status. The browser showed a not-found page, but anything reading the raw response saw a perfectly healthy homepage. That fooled search engines, and it fooled uptime monitors too: a keyword check against a missing checkmeup.net page passed, because the homepage happens to mention \"A, AAAA, CNAME\" and the keyword was \"AAA\". Missing pages now get a real 404.",
  },
  {
    type: 'h2',
    text: 'Changing an interval now takes effect',
  },
  {
    type: 'p',
    text: "Switching an uptime monitor from every 10 minutes to every minute didn't speed anything up until the old 10 minutes ran out. A shorter interval now applies right away, and a longer one takes over after the check that's already scheduled. Cron monitors had the mirror-image bug, and a worse one: moving a job to a less frequent schedule could fire a false \"missed its ping\". A schedule change now gives the job one full new window.",
  },
  {
    type: 'h2',
    text: 'Also these releases',
  },
  {
    type: 'ul',
    items: [
      'Release notes like this one moved from /blog to a new /changelog page and are kept out of search results. The blog is for articles meant to be found; release notes are for people already following along.',
      'Every page now has one URL: a trailing-slash version redirects to the plain one, so the same page no longer shows up twice in search results.',
      'Signal and Microsoft Teams alerts are off the roadmap. That decision, and the reasons for it, are written up in the alerting decision record.',
      'Behind the scenes: one shared renderer now builds every down and recovery alert across all monitor types, checked byte for byte against the old output. The development environment moved from a devcontainer to plain pinned tools, and CI runs with read-only repository access.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "The next release is a big visual one: a redesign of every screen. Full commit history and every architecture decision record are on GitHub if you want the unfiltered version.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
