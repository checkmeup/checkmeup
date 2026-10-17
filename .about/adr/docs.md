# Where the project's knowledge lives

Date: 2026-10-17

Decisions are ADRs in `.about/adr/`, one file per subject, saying what holds now; a changed decision is edited in place and the dropped choice moves to Considered options. Features are PRDs in `.about/prd/`, one per feature, with acceptance criteria linked to the tests that check them. Domain terms are in `.about/glossary.md`. Anything open — an undecided question, a finding left for later — is an item in `.about/inbox.md` until something settles it. The baloo plugin's skills (adr, prd, glossary, inbox) keep these files.

`docs/roadmap.md` is a Now / Next / Later list of PRDs with no dates. `docs/incidents/` holds one file per production incident on Checkmeup's own infrastructure: what broke, whether our monitors caught it, impact, cause, follow-up. `docs/reference/` is living how-to and `docs/proposals/` holds feature ideas not yet picked up. `docs/knowledge/mvp-history.md` freezes the MVP build plan. How the code works is read from the code; why it is built that way is in the ADRs.

There are no monthly reports and no hours log: what shipped is in the git log and the GitHub Releases changelog.

## Considered options

- Numbered ADRs in `docs/decisions/` and epics with user stories in `docs/stories/` (until 2026-10-13): moved to `.about/` so baloo's skills and its session review work on them; the numbered ADRs were merged into one file per subject.
- Monthly reports in `docs/reports/` (until 2026-10-13): stopped along with hours logging; they repeated the git log and the changelog. The June–September 2026 reports were deleted on 2026-10-14; git keeps them.
- Architecture snapshots in `docs/knowledge/` and a `docs/investigations/` folder (until 2026-10-14): the snapshots restated the code with file:line references that went stale, and their decisions are in the ADRs; no investigation was ever written.
- A tech-debt list in `docs/reference/tech-debts.md` (until 2026-10-14): a debt left for later is an item in `.about/inbox.md`.
- A generated code map in `docs/codemap/` (until 2026-10-14): a one-off snapshot nothing regenerated.
- A dated, hour-budgeted roadmap: fitted the MVP sprint, not reactive post-launch work.
- An hours log in `docs/reports/hours/`, kept by the `log-hours` skill, with a hook that blocked `gh pr create` until hours were logged (until 2026-10-13): dropped entirely; git keeps the logs.
- A `session_summary.py` hook proposing ADR and doc updates at session end (until 2026-10-13): replaced by baloo's session review.
