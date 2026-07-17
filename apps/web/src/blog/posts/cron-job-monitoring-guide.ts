import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "A cron job doesn't usually announce its own death. It doesn't throw an exception into your error tracker, it doesn't page anyone, it doesn't even exit non-zero half the time — it just stops appearing in the log, and nothing downstream is looking for its absence. The backup that quietly stopped running three weeks ago, the invoice job that's been skipping every other client since a dependency bump, the cleanup script that died the day the server rebooted and the crontab didn't survive it — these are the failures that don't show up until someone asks for a backup that isn't there.",
  },
  {
    type: 'h2',
    text: 'What is cron job monitoring?',
  },
  {
    type: 'p',
    text: "Cron job monitoring answers a question your crontab can't: did this job actually run? It works as a dead man's switch, inverted from how most monitoring works. An uptime check calls your server and expects an answer; a cron monitor waits for your job to call it, and raises an alarm if the call doesn't come. The job itself reports success — the monitor's only role is noticing when a report was expected and didn't arrive.",
  },
  {
    type: 'h2',
    text: 'Why cron jobs fail silently',
  },
  {
    type: 'p',
    text: 'Cron itself has no concept of "this should have run" — it fires the schedule and moves on, whether or not the job existed to receive it. A handful of ordinary events can break a job without ever producing an error anyone sees:',
  },
  {
    type: 'ul',
    items: [
      'A server reboot (patching, a host migration) and the crontab entry never made it back, or came back with the wrong PATH',
      'A dependency, API key, or credential the script relies on expired or changed, and the script exits early — sometimes with status 0',
      "cron's own mail delivery is unconfigured or going to an inbox nobody reads, so a script that does error correctly is still invisible",
      'The previous run is still stuck (mid-download, waiting on a lock) and the schedule just skips the next one',
      'A deploy overwrote or commented out the crontab line and nobody noticed because nothing broke immediately',
    ],
  },
  {
    type: 'p',
    text: "None of these look like an incident from the server's point of view. The only way to catch them is to expect a signal on a schedule and notice when it stops arriving — which is exactly what a cron monitor is built to do, and cron itself is not.",
  },
  {
    type: 'h2',
    text: 'Setting up cron job monitoring in three steps',
  },
  {
    type: 'h3',
    text: '1. Pick a grace period that matches how the job actually behaves',
  },
  {
    type: 'p',
    text: "A grace period is how long the monitor waits past the expected time before calling a run missed. Too tight, and normal jitter — a slow disk, a busy host, a job that runs 90 seconds long instead of 60 — trips a false alarm and trains you to ignore alerts. Too loose, and a real failure sits unnoticed for hours. Base it on the job's own variance: watch a few real runs first if you don't already know how long it takes end to end, and set the grace period a bit above the slowest one you've seen, not the average.",
  },
  {
    type: 'h3',
    text: '2. Call a URL at the end of the job, not the start',
  },
  {
    type: 'p',
    text: "The monitor only knows a job succeeded if the job tells it so — chain the ping onto the command that already runs, after the real work, so a failing script never gets the chance to report success it didn't earn:",
  },
  {
    type: 'code',
    lang: 'bash',
    text: '# in your crontab or script\n0 3 * * * /opt/scripts/nightly-backup.sh && curl -fsS https://checkmeup.net/ping/YOUR_TOKEN',
  },
  {
    type: 'p',
    text: '`&&` is doing the important work here: curl only fires if the backup script exited 0. A script that dies partway through never sends the ping, which is exactly the failure the monitor exists to catch. `-f` also makes curl itself fail loudly on a bad response instead of silently swallowing it.',
  },
  {
    type: 'h3',
    text: "3. Wire up an alert channel you'll actually see in time",
  },
  {
    type: 'p',
    text: "An alert that lands in an inbox nobody checks until Monday isn't meaningfully different from no alert at all. Pick something you'll notice within the grace period that actually matters for that job — a chat app you already have open, an SMS for anything client-facing, a webhook into whatever on-call system you already run. The channel matters as much as the check.",
  },
  {
    type: 'h2',
    text: 'What to check before trusting a tool with it',
  },
  {
    type: 'ul',
    items: [
      "An unguessable ping URL — it's the only credential protecting the endpoint, so it needs to not be a sequential ID",
      'A per-job grace period, not one global timeout for every schedule you run',
      'An execution log you can actually look at — not just a current status badge, but a history of when pings did and did not arrive',
      "A monitoring endpoint that always returns 200, so a blip in the monitoring service can't fail a job that's actually fine",
      'More than one alert channel, so a single dead integration (an expired bot token, a bounced email) never becomes a silent monitoring gap on top of the silent job failure it was supposed to catch',
    ],
  },
  {
    type: 'h2',
    text: 'The mistake that undoes all of this',
  },
  {
    type: 'p',
    text: "Setting up monitoring and never simulating a failure is the most common way this quietly stops working. Comment out the cron line for a minute, or send a bad exit code on purpose, and confirm the alert actually arrives, on the channel you expect, within the grace period you set. A monitor nobody has ever seen fire is a monitor nobody actually knows works — the same blind spot as the cron job it's meant to be watching.",
  },
  {
    type: 'divider',
  },
  {
    type: 'p',
    text: "Checkmeup's cron monitor is exactly the dead man's switch described above — unique ping URL, per-job grace period, execution and incident logs, alerts over Telegram, email, SMS, or webhook. The Hobby plan is free for up to 10 monitors, no credit card required, if you want to try it on the job you're least sure is still running.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
