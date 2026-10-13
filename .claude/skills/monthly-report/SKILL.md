---
name: monthly-report
description: Build or update this month's docs/reports/YYYY-MM.md snapshot — a Shipped log of merged work. Use when asked to "update the monthly report" or "add this to the report".
---

# Monthly report

`docs/reports/YYYY-MM.md` is a monthly snapshot of what shipped — see
`docs/reports/2026-07.md` for the current format. It complements
`docs/decisions/` (ADR log) rather than duplicating it.

## Steps

**1. If the file doesn't exist yet for the current month**, create it
from the previous month's structure:

```markdown
# <Month> <Year>

## Shipped
```

**2. Add a bullet to `## Shipped`** per notable merged item — a shipped
epic, an ADR, a redesign, a meaningful docs pass, a notable fix. Not
every commit needs an entry; batch related commits into one bullet.
Each bullet:

- Links the epic/ADR if one exists: `[EP-33](../stories/ep-33-...md)` or
  `[ADR-025](../decisions/025-...md)`
- States the date it shipped (`— Jul 1`)
- Explains **what** shipped and, more importantly, **why** it was built
  that way — the standout design choice, a scope simplification and the
  reason, or a bug caught and how (see `docs/reports/2026-07.md`'s Port
  monitoring and Codacy-cleanup entries for the level of detail expected)
- Is one dense paragraph, not a multi-bullet breakdown

**3. Don't re-derive from git blindly** — cross-check against
`docs/stories/backlog.md`'s "stories done" fractions and recently-added
`docs/decisions/` files, since those are the authoritative record of
what actually shipped vs. what's still in progress.
