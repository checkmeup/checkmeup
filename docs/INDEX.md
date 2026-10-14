# Docs Index

<!-- AI LOAD ORDER: INDEX.md → ../.about/ (adr, prd, glossary, inbox) → reference/ → proposals/ -->

Navigation cache for `docs/`. Always load this file first. Decisions, feature requirements, domain terms and open questions live outside `docs/`, in [`.about/`](../.about/) — see below. Every other doc carries its own
frontmatter (`type`, `status`, `updated`, `tags`).

**Folder roles:**

| Folder | Question answered | Mutable? |
|---|---|---|
| `reference/` | How does the system work right now? | Yes — update when code or ops change |
| `knowledge/` | How was it built? (the MVP history) | No — frozen |
| `proposals/` | What might we add or change? | Yes — until decided |
| `incidents/` | What broke in production? | No — permanent once written |

**Archive rule:** Set `status: archived` in frontmatter. Remove from the tables below.
Do not move files — lifecycle is metadata, not filesystem position.

**Controlled tag vocabulary** (reuse these, don't invent new ones): `billing`, `paddle`, `twilio`,
`ops`, `deployment`, `kamal`, `hetzner`, `design`, `tokens`, `colors`, `logo`, `security`, `dos`,
`audit`, `architecture`, `maintainability`, `backend`, `frontend`, `vue3`, `auth`, `monitoring`,
`cron`, `uptime`, `ssl`, `domain`, `port`, `dns`, `alerts`, `telegram`, `sms`, `email`, `webhook`,
`status-page`, `features`, `competitors`, `backlog`, `history`, `planning`.

## reference/ — living, update-on-change

| Doc | Status | Hook |
| --- | --- | --- |
| [billing-setup.md](reference/billing-setup.md) | active | Paddle dashboard checklist to activate billing |
| [deploy.md](reference/deploy.md) | active | Kamal deployment guide, cost breakdown, gotchas |
| [design.md](reference/design.md) | active | Color tokens, logo assets, theme conventions |
| [limits.md](reference/limits.md) | active | DoS / overload vulnerability audit (all findings resolved) |
| [twilio-setup.md](reference/twilio-setup.md) | active | Twilio account setup checklist for SMS alerts |

## .about/ — decisions, requirements, terms, open questions

Kept by the baloo plugin's skills ([ADR docs](../.about/adr/docs.md)).

| Where | What | Kept by |
| --- | --- | --- |
| [.about/adr/](../.about/adr/) | Decisions as they stand now, one file per subject (workers, limits, billing, alerting, auth, database, api, status-pages, infrastructure, testing, product, frontend, seo, dev-environment, docs) | `baloo:adr` |
| [.about/prd/](../.about/prd/) | One PRD per feature; acceptance criteria link the tests that check them | `baloo:prd` |
| [.about/glossary.md](../.about/glossary.md) | Domain terms | `baloo:glossary` |
| [.about/inbox.md](../.about/inbox.md) | Open questions and findings left for later | `baloo:inbox` |

## knowledge/ — frozen history

| Doc | Status | Hook |
| --- | --- | --- |
| [mvp-history.md](knowledge/mvp-history.md) | current | Frozen record of the MVP build: phases, hours, decisions |

## proposals/ — undecided options only

| Doc | Status | Hook |
| --- | --- | --- |
| [bucket-list.md](proposals/bucket-list.md) | proposed | Competitor gap analysis — unprioritized feature ideas |

## incidents/ — production incidents

See [incidents/README.md](incidents/README.md) for the template. One file per incident.

## Root-level docs

| Doc | Hook |
| --- | --- |
| [roadmap.md](roadmap.md) | Now / Next / Later product work |

## Conventions

- **Naming:** kebab-case, descriptive noun phrase, no dates in the filename (dates live in `updated`).
- **One topic per file:** if a doc grows past a few hundred lines and covers more than one thing
  someone would search for independently, split it.
- **Status drives visibility:** only `active` / `current` / `proposed` docs appear in the tables above.
  Archived docs stay in place and are excluded from active reasoning.
- **Standing "don't do X" rules** belong in `CLAUDE.md`, not here.
