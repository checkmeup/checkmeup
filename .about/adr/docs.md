# Where the project's knowledge lives

Date: 2026-10-13

Decisions are ADRs in `.about/adr/`, one file per subject, saying what holds now; a changed decision is edited in place and the dropped choice moves to Considered options. Features are PRDs in `.about/prd/`, one per feature, with acceptance criteria linked to the tests that check them. Domain terms are in `.about/glossary.md`. Anything open — an undecided question, a finding left for later — is an item in `.about/inbox.md` until something settles it. The baloo plugin's skills (adr, prd, glossary, inbox) keep these files.

`docs/roadmap.md` is a Now / Next / Later list of PRDs with no dates. `docs/incidents/` holds one file per production incident on Checkmeup's own infrastructure: what broke, whether our monitors caught it, impact, cause, follow-up. `docs/reference/` is living how-to, `docs/knowledge/` architectural snapshots, `docs/investigations/` and `docs/proposals/` one-off write-ups. `docs/knowledge/mvp-history.md` freezes the MVP build plan.

There are no monthly reports: what shipped is in the git log and the GitHub Releases changelog. `docs/reports/` keeps the June–September 2026 reports as history and gets no new ones.

## Considered options

- Numbered ADRs in `docs/decisions/` and epics with user stories in `docs/stories/` (until 2026-10-13): moved to `.about/` so baloo's skills and its session review work on them; the numbered ADRs were merged into one file per subject.
- Monthly reports in `docs/reports/` (until 2026-10-13): stopped along with hours logging; they repeated the git log and the changelog.
- A dated, hour-budgeted roadmap: fitted the MVP sprint, not reactive post-launch work.
