---
name: log-hours
description: Log hours worked for a day into docs/hours.md and roll the new total into that month's docs/reports/YYYY-MM.md Notes section, based on that day's git commit history. Use when asked to "log hours", "log today's hours", "update hours for <date>", or "log hours for yesterday/Thursday/etc".
---

# Log hours

Reconstructs a day's effort from git history rather than asked-for
estimates, per CLAUDE.md's hours-logging convention.

A `PreToolUse` hook (`.claude/hooks/require_log_hours.py`) blocks
`gh pr create` until `docs/hours.md` has been touched by a commit on the
current branch — hours ship in the *same* PR as the work, not a
follow-up `docs/hours-*` PR. Run this skill and commit the result on the
current branch before creating the PR, not after.

## Steps

**1. Resolve the date.** Convert any relative reference ("today",
"yesterday", "Thursday") to an absolute `YYYY-MM-DD` — the log must stay
interpretable long after the conversation ends.

**2. Pull that day's commits.**

```bash
git log --all --since="<date> 00:00" --until="<date> 23:59" --pretty=format:'%h %s' --date=short
```

Using `--all` surfaces stash refs too — **exclude** any entry titled
`On <branch>:` / `index on <branch>:` / `untracked files on <branch>:`,
those are `refs/stash` artifacts, not real commits.

**3. Group commits into logical tasks, not one line per commit.** Small
related commits (a fix-up, a lint pass, a follow-up typo fix on the same
feature) collapse into one line. Estimate effort from diff size and
complexity — **quarter-hour increments** (0.25/0.5/0.75/1/1.25 h, etc.),
minimum 0.25 h per task/line. Round each line to the nearest 15 minutes;
don't stack unrelated small tasks into one inflated line just to avoid a
short one — a genuinely 15-minute fix gets its own 0.25 h row.

**4. Include non-commit work from the same session** if mentioned in
conversation — copywriting, launch/marketing text, doc corrections made
directly without a commit — each gets its own line.

**5. Append to `docs/hours.md`'s Log table**, matching the exact existing
row format:

```text
| Date       | Day | Epic/Story                                      | Hours |
| ---------- | --- | ----------------------------------------------- | ----- |
| YYYY-MM-DD | Mon | <task description, matching commit's own wording style> | N h   |
```

`N` is a quarter-hour value — `1 h`, `1.25 h`, `1.5 h`, `1.75 h`, `2 h`,
etc., never anything finer than 0.25. Reference `EP-XX`/`US-XXXX` in the
description where the work maps to a tracked story; otherwise describe
the work directly (e.g. "Codacy Stylelint fixes (style.css)").

**6. Roll the day's total into `docs/reports/YYYY-MM.md`'s `## Notes`
section** — update both the month's total and the cumulative
since-project-start total:

```text
**Total effort this month: N h**. Day-by-day breakdown in [hours.md](../hours.md) — not duplicated here.

**Cumulative total since project start: N h** across N logged days — average N.NN h/day.
```

Recompute both totals from `docs/hours.md` rather than just adding —
cheapest way to catch a prior arithmetic slip. The rolled-up totals and
the average-per-day figure are derived arithmetic and can land on any
decimal (e.g. "9.55 h/day"), unlike the per-row Hours column, which
stays on quarter-hour steps.

## Notes

- Post-MVP work (`ADR-022`) intentionally has no hour *estimates*, only
  logged actuals — don't add an estimate column.
- If today's report file (`docs/reports/YYYY-MM.md`) doesn't exist yet
  for a new month, look at the previous month's file for the section
  structure (`## Shipped`, `## Notes`) before creating it.
