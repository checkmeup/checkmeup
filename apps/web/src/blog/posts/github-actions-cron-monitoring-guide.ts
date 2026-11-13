import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "The nightly sync workflow has a green checkmark on its last run. The problem is that the last run was in August. Nobody pushed to the repository for two months, GitHub quietly disabled the schedule, and the only notice went to the repo owner's email among a few hundred other GitHub notifications. There was no red X, no failed job, no alert — the workflow didn't break, it stopped existing, and a workflow that never starts can't fail in a way your CI dashboard will show you.",
  },
  {
    type: 'h2',
    text: 'Why a scheduled workflow stops without failing',
  },
  {
    type: 'p',
    text: "A schedule trigger is the one part of GitHub Actions that isn't driven by something you did. Nothing in your repository goes red when it stops firing, and there are several ordinary ways for that to happen:",
  },
  {
    type: 'ul',
    items: [
      'GitHub disables scheduled workflows in a public repository after 60 days without commits, and emails the owner about it — a repository that is finished rather than abandoned hits this regularly',
      'A schedule trigger only runs from the default branch, so a workflow file that lives on a branch, or gets renamed during a merge, is never scheduled at all',
      'Scheduled runs go into a shared queue and can be delayed by several minutes, or dropped entirely, when the whole platform is busy — the top of the hour is the worst slot for exactly this reason',
      'Anyone can disable a workflow in the Actions tab in one click, and nothing afterwards hints that it is off',
      'On a private repository, a spending limit reached mid-month stops runs from starting, while the workflow file still looks perfectly healthy',
    ],
  },
  {
    type: 'p',
    text: "What these have in common is that the failure is an absence. Error tracking, a failed-job notification, a Slack integration on the repository — all of them report on runs that happened. None of them has anything to say about the run that didn't.",
  },
  {
    type: 'h2',
    text: 'What a ping-based check actually proves (and what it does not)',
  },
  {
    type: 'p',
    text: "Monitoring a schedule inverts the usual arrangement: instead of something calling your workflow, your workflow calls the monitor at the end of a successful run. The monitor knows the schedule you told it about and a grace period; when the expected call doesn't arrive in time, it opens an incident and alerts you. That's a dead man's switch, and it's the only mechanism that treats \"nothing happened\" as an event.",
  },
  {
    type: 'p',
    text: "It's worth being precise about what that proves, because it is less than it looks like. A ping means one thing: the last step of the job ran on a runner and exited 0. It does not mean the sync actually copied anything, that the backup it uploaded is restorable, or that the data it wrote is correct. A workflow that silently processes zero rows and exits cleanly pings exactly like a healthy one. If correctness matters more than liveness, the ping is the floor, not the ceiling — put an assertion in the job itself, so that a bad outcome makes the step fail and the ping never gets sent.",
  },
  {
    type: 'p',
    text: 'The other half of the honesty: a missed run is caught by the grace period expiring, not instantly. There is no failure signal to send — a job that dies halfway through simply never reaches its ping step, and you hear about it when the monitor gives up waiting. That delay is a setting you choose, which makes the next section the one that matters.',
  },
  {
    type: 'h2',
    text: 'Setting it up in three steps',
  },
  {
    type: 'h3',
    text: "1. Set the grace period to GitHub's queue, not your cron line",
  },
  {
    type: 'p',
    text: "This is where most of these monitors go wrong. The cron expression in your workflow is when GitHub starts thinking about the run, not when it starts. Queue delays of five to fifteen minutes are normal at busy times, and the job's own runtime comes on top of that. Set the monitor's schedule to the same cron expression the workflow uses, then give it a grace period that covers the worst queue delay plus the slowest run you've actually seen — for a nightly job, 30 to 60 minutes is realistic. If you set five minutes, you'll be chasing GitHub's queue, and you'll start ignoring the alerts within a week.",
  },
  {
    type: 'h3',
    text: '2. Make the ping the last step, and pass the run with it',
  },
  {
    type: 'p',
    text: 'A step in GitHub Actions only runs if every step before it succeeded, so a ping as the final step is already conditional on the real work having passed — no if: condition needed:',
  },
  {
    type: 'code',
    lang: 'yaml',
    text: 'name: nightly-sync\n\non:\n  schedule:\n    - cron: \'0 3 * * *\'\n  workflow_dispatch:\n\njobs:\n  sync:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - name: Sync\n        run: ./scripts/sync.sh\n      - name: Report success\n        run: >\n          curl -fsS --retry 3\n          "https://checkmeup.net/ping/${{ secrets.CHECKMEUP_PING_TOKEN }}?run=${{ github.run_id }}&attempt=${{ github.run_attempt }}"',
  },
  {
    type: 'p',
    text: "The token goes in a repository secret rather than the workflow file — it's the only thing protecting the endpoint, and workflow files are readable by anyone who can see the repository. The query string is recorded with the ping, so when an alert does fire you can open the run it refers to instead of guessing which one it was. The --retry 3 flag covers a dropped connection on the runner's way out; the endpoint answers 200 whether or not anything was recorded, so a bad minute on the monitoring side can never fail a job that's fine.",
  },
  {
    type: 'h3',
    text: '3. For a long workflow, signal the start as well',
  },
  {
    type: 'p',
    text: 'A job that starts and then hangs — a runner waiting on a lock, a download that never finishes, a step with no timeout — is invisible to a completion ping until the whole grace period has run out, which for a nightly job may be the next morning. Pinging a start URL as the first step opens a run the monitor can time, and a maximum duration on the monitor turns "started at 03:00, still not finished at 04:00" into an alert of its own:',
  },
  {
    type: 'code',
    lang: 'yaml',
    text: '      - name: Report start\n        run: curl -fsS "https://checkmeup.net/ping/${{ secrets.CHECKMEUP_PING_TOKEN }}/start"',
  },
  {
    type: 'p',
    text: "GitHub's own timeout-minutes setting on the job does a related thing — it cancels the run — but a cancelled run is still a run that produced nothing, and the monitor is what tells you it happened.",
  },
  {
    type: 'h2',
    text: 'What to look for in a monitoring tool',
  },
  {
    type: 'ul',
    items: [
      "An unguessable ping URL, since it's the only credential on the endpoint, and one you can rotate without editing every workflow that uses it",
      'A per-job grace period, so a nightly sync and a five-minute health job are not forced onto the same tolerance',
      'An endpoint that always returns 200 and never fails your pipeline because the monitoring service is having a bad day',
      'A ping history you can read — when pings arrived, how late they were, which run each one came from — rather than only a current up/down badge',
      "More than one alert channel, so an expired bot token doesn't turn into a silent monitoring gap sitting on top of a silent job failure",
    ],
  },
  {
    type: 'h2',
    text: 'The mistake worth avoiding',
  },
  {
    type: 'p',
    text: "Don't put if: always() or continue-on-error: true on the ping step. It looks like defensive engineering — the ping fires no matter what, the monitor stays green, nobody gets woken up. That is precisely the problem: the monitor now reports that the workflow ran, which you already knew from the Actions tab, and says nothing about whether it worked. The ping is only worth having because a failing job can't send it. Set it up, then prove it: disable the workflow in the Actions tab for one scheduled slot and check that the alert arrives on the channel you expect, within the grace period you chose. A monitor nobody has ever seen fire is just another workflow you believe is running.",
  },
  {
    type: 'divider',
  },
  {
    type: 'p',
    text: "Checkmeup's cron monitor is the dead man's switch described above: a cron schedule with its own grace period, an unguessable ping URL that always answers 200, an optional start ping with a maximum duration for jobs that hang, the query string of each ping kept with it, and alerts over Telegram, Slack, email, SMS or webhook — with an option to alert only after several misses in a row, for a schedule that's allowed the occasional slip. The Hobby plan is free for up to 10 monitors, no credit card, which is enough for every scheduled workflow most repositories have.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
