---
name: monthly-report
description: Build or update this month's docs/reports/YYYY-MM.md snapshot — a Shipped log of merged work plus a Notes section with hours totals. Use when asked to "update the monthly report", "add this to the report", or when the log-hours skill needs a report file that doesn't exist yet for the current month.
---

# Monthly report

`docs/reports/YYYY-MM.md` is a monthly snapshot of what shipped and how
much effort it took — see `docs/reports/2026-07.md` for the current
format. It complements `docs/hours.md` (daily log) and `docs/decisions/`
(ADR log) rather than duplicating either.

## Steps

**1. If the file doesn't exist yet for the current month**, create it
from the previous month's structure:

```markdown
# <Month> <Year>

## Shipped

## Notes

**Total effort this month: 0 h**. Day-by-day breakdown in [hours.md](../hours.md) — not duplicated here.

**Cumulative total since project start: N h** across N logged days — average N.NN h/day.
```

**2. Add a bullet to `## Shipped`** per notable merged item — a shipped
epic, an ADR, a redesign, a meaningful docs pass, a notable fix. Not
every commit needs an entry; batch related commits into one bullet, the
same grouping judgment as the log-hours skill. Each bullet:

- Links the epic/ADR if one exists: `[EP-33](../stories/ep-33-...md)` or
  `[ADR-025](../decisions/025-...md)`
- States the date it shipped (`— Jul 1`)
- Explains **what** shipped and, more importantly, **why** it was built
  that way — the standout design choice, a scope simplification and the
  reason, or a bug caught and how (see `docs/reports/2026-07.md`'s Port
  monitoring and Codacy-cleanup entries for the level of detail expected)
- Is one dense paragraph, not a multi-bullet breakdown

**3. Update `## Notes`** — this is where the log-hours skill rolls in its daily
total; if you're writing this section standalone, recompute both numbers
from `docs/hours.md` rather than incrementing by feel:

```bash
grep -c '^| <YYYY-MM>' docs/hours.md   # days logged this month, e.g. '^| 2026-07'
```

**4. Don't re-derive from git blindly** — cross-check against
`docs/stories/backlog.md`'s "stories done" fractions and recently-added
`docs/decisions/` files, since those are the authoritative record of
what actually shipped vs. what's still in progress.
