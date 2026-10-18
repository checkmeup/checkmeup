# Product pages per monitor type and "alternative to" pages

## Problem

Search Console shows almost no non-brand search traffic (Inbox "Almost no organic search traffic"). People searching "cron job monitoring" or "uptime monitoring" find only blog guides here, with no page that says what Checkmeup does for that need, what it costs and how to start. People already choosing between tools search "Healthchecks.io alternative", "UptimeRobot alternative" or "Cronitor alternative", and nothing here answers them; one combined comparison post covers all three.

## Users

Developers, freelancers and agencies searching for a monitoring tool for one need, or comparing Checkmeup with a tool they use or are considering.

## Success signal

By 2026-12-31, Search Console shows impressions for non-brand queries on these pages (baseline 0), and GA4 counts at least one `sign_up` key event from a session that landed on one of them. Both are measured with what already exists: Search Console's per-page report and GA4's landing page dimension.

## Scope

- One product page per monitor type at `/cron-monitoring`, `/uptime-monitoring`, `/ssl-monitoring`, `/domain-expiry-monitoring`, `/port-monitoring` and `/dns-monitoring`: what it watches, how the check works, what a passing check doesn't prove, the alert channels, the plans, a sign-up button, and links to the type's guide or free checker. Every claim matches the type's PRD.
- One page per competitor at `/alternatives/healthchecks-io`, `/alternatives/uptimerobot` and `/alternatives/cronitor`: a side-by-side table, reasons to switch, when the competitor is the better choice, how to switch, and the sources of every competitor fact, dated.
- All nine pages prerendered with their own title, description and `<h1>` (ADR seo), listed in `sitemap.xml`, and linked from the site footer.

## Non-goals

- An `/alternatives` index page, or pages for other competitors.
- Importing monitors from a competitor.
- Keeping competitor facts current automatically: each page shows the date they were checked.

## Acceptance criteria

- Each monitor type has its page at its own path, with its own `<h1>`, title and canonical URL.
  → test: apps/web/src/views/MonitorTypeView.test.ts "serves the $type page at $path"
  → test: apps/web/src/views/MonitorTypeView.test.ts "gives the $type page its own heading, title and canonical URL"
- Each monitor type page links to sign-up and to its guide or free checker.
  → test: apps/web/src/views/MonitorTypeView.test.ts "links the $type page to sign-up and to $link"
- Each monitor type page shows every plan with its price and says what its check doesn't prove.
  → test: apps/web/src/views/MonitorTypeView.test.ts "shows every plan with its price, and says what its check does not prove"
- Every monitor type page is linked from the footer of every marketing page.
  → test: apps/web/src/layouts/LandingLayout.test.ts "links every monitor type page from the footer"
- Each competitor has its page at `/alternatives/<slug>`, with its own `<h1>`, title and canonical URL.
  → test: apps/web/src/views/AlternativeView.test.ts "serves the $name page at /alternatives/$slug"
  → test: apps/web/src/views/AlternativeView.test.ts "gives the $name page its own heading, title and canonical URL"
- Each alternative page compares Checkmeup and the competitor side by side, shows the date its facts were checked, and links its sources.
  → test: apps/web/src/views/AlternativeView.test.ts "compares Checkmeup with $name side by side, dated, with its pricing page as a source"
- Each alternative page says when the competitor is the better choice, and links to sign-up.
  → test: apps/web/src/views/AlternativeView.test.ts "says when $name is the better choice, and links to sign-up"
- Every alternative page is linked from the footer of every marketing page.
  → test: apps/web/src/layouts/LandingLayout.test.ts "links each alternative page from the footer"
- All nine pages are prerendered with their own `<h1>`, title and description, and listed in `sitemap.xml`.
