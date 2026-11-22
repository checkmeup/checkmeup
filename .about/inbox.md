# Inbox

## Architecture deepening candidates

2026-10-14 · session review · kept 2026-11-13

An architecture review found places where each monitor type repeats logic that should exist once. One is left: SSL and domain as one "Expiry monitor" (worth exploring), as their workers, handlers and edit views differ only slightly; merging them would need a glossary term.

The other two were taken out on 2026-11-22: the plan checks copied across the monitor handlers (`handler/monitor_limits.go`), and the incident lifecycle uptime, port and DNS each ran on their own (`worker/incident_tracker.go`). What the product waits on now is traffic rather than code. Settled by agreeing a plan for the Expiry monitor, or by dropping it.

## Free tools without sign-up

2026-10-17 · traffic review · kept 2026-11-22

Public, no-account checkers people search for: SSL certificate, domain expiry (RDAP/WHOIS), DNS lookup, TCP port, cron expression explainer. The backend already has `rdap`, `whois` and the worker's checks; each result offers to monitor it automatically. Also linkable assets for directories and Show HN (see "Launch and directory listings"). Needs per-IP rate limits and SSRF-safe targets. SSL and domain expiry checkers came first (PRD free-tools), the DNS lookup followed on 2026-10-31 (`/tools/dns-lookup`), and the cron explainer on 2026-11-21 (`/tools/cron-expression`), the one that leads to cron monitoring, the product's core. Only the TCP port checker remains, a non-goal of PRD free-tools for now, waiting on the tools' 2026-12-15 success signal. Settled when it is built or dropped.

## Stack-specific cron guides

2026-10-17 · traffic review · kept 2026-11-22

Long-tail guides: monitoring a Kubernetes CronJob, a GitHub Actions schedule, Laravel's scheduler, systemd timers, Celery beat. Pairs with the CLI wrapper in `docs/proposals/bucket-list.md`. The GitHub Actions one shipped on 2026-11-13 (`/blog/github-actions-cron-monitoring-guide`), picked because the head term "cron job monitoring" sits at position 73.6 while the tail is open, and because GitHub disabling a schedule after 60 days without commits is a failure the product's own dead man's switch catches. The other four wait on what that one does: whether it takes impressions, and on which queries. Settled when the stacks to cover are picked and written, or the idea is dropped.

## Launch and directory listings

2026-10-17 · traffic review · kept 2026-11-22

Channels only the founder can post to, for the outside signals a young domain needs to rank: almost no clicks come from search yet (Search Console and `mise run traffic` hold the numbers). Done: awesome-monitoring listed ([PR #121](https://github.com/Enapiuz/awesome-monitoring/pull/121), merged 2026-10-30); AlternativeTo and SaaSHub pages filled in, SaaSHub's "Product of the Day" judged not worth the weeks of voting; awesome-status-pages is closed to Checkmeup (CLAUDE.md). Left to do:

- awesome-devops: reopen [PR #636](https://github.com/wmariuss/awesome-devops/pull/636) after 2026-12-12, when `checkmeup/checkmeup` turns 6 months old (`gh pr reopen 636 -R wmariuss/awesome-devops`, then an empty push to rerun the check).
- Show HN: [item 50235049](https://news.ycombinator.com/item?id=50235049) stays `dead` after asking `hn@ycombinator.com` to unflag it; repost after a few weeks of commenting from the founder's account, on a Tuesday–Thursday around 8–10am US Eastern, without changing the text. HN and Reddit gate posting on karma, and generated comments are against HN's rules, so Claude doesn't comment.
- Cross-posts with a canonical link to the blog, one every few days, each marked AI-assisted as [DEV's guidelines](https://dev.to/guidelines-for-ai-assisted-articles-on-dev) require (DEV's AI Disclosure set to "Some AI" and a closing line, added to the earlier ones on 2026-11-19): dev.to has RDAP vs WHOIS, the SSL story, the cron guide and, since 2026-11-19, the domain expiry post, the first with a cover image; "Three Days to MVP" is next; [checkmeup.hashnode.dev](https://checkmeup.hashnode.dev) has the same four, with the same closing line, covers on the cron guide and the domain expiry post, and the publication's description, logos and social image filled in (2026-11-19); Hashnode drops a cover set on a draft when it is published, so a cover goes on after publishing, through Update. Since 2026-11-22 every page has its own social image (`https://checkmeup.net/img/og/<page path>.png`, PRD social-images), which can serve as a post's cover on both. No visit has come from either yet. HackerNoon is dropped: it rejected the GitHub Actions guide on 2026-11-18 because a canonical link to checkmeup.net makes a post brand content, published only for a paid brand credit.
- Product Hunt, Reddit (r/webdev, r/selfhosted, r/freelance), Indie Hackers: not tried.

Settled when the channels are chosen and tried.
