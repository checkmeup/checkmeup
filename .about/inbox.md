## Move the project's docs into .about/

2026-10-13 · session review

ADRs still live in `docs/decisions/` (numbered, scaffolded by `new-adr`), stories in `docs/stories/` (`new-story`) and open questions in `docs/decisions/backlog.md`. baloo reads `.about/adr/`, `.about/prd/` and `.about/inbox.md`. The agreed PR moves the ADRs to `.about/adr/`, the stories to PRDs and `backlog.md` into this Inbox, deletes the `new-adr` and `new-story` skills, and fixes every link. Settled when that PR merges.

## Check Search Console once the changelog branch ships

2026-10-13 · session review

`feat/changelog-noindex-release-notes` (release notes `noindex`, `/changelog`, trailing-slash 301, path-only canonical) isn't merged yet. Google drops pages only after a recrawl, which can take weeks. The "Page with redirect" count should grow as trailing-slash variants become 301s. The guess that today's 4 redirect pages are the `http://` and `www.` variants was never checked. Settled when Search Console shows the release notes excluded by `noindex`, no trailing-slash duplicates indexed, and the redirect pages accounted for.

## Are monthly reports still written?

2026-10-13 · session review

The `monthly-report` skill was deleted, and the old reports in `docs/reports/` were kept as history. ADR-022, `CLAUDE.md` ("monthly snapshots in `docs/reports/`") and `docs/INDEX.md` ("monthly shipping logs") still describe monthly reports as an ongoing practice. Settled by deciding whether they continue. If not, mark `docs/reports/` as history in those three files. If so, say how a report gets written without the skill.
