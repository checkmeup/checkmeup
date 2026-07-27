---
name: new-adr
description: Scaffold a new Architecture Decision Record under docs/decisions/, numbered sequentially, in this repo's actual ADR format. Use when asked to "write an ADR", "document this decision", "create ADR for X", or after making a non-obvious architectural choice worth recording.
---

# New ADR

Records a decision under `docs/decisions/`, following the format the
existing ADRs actually use — **not** `_template.md`'s YAML-frontmatter
shape, which no real ADR in this repo follows. Match the files, not the
template.

## Steps

**1. Find the next number.**

```bash
ls docs/decisions | grep -E '^[0-9]{3}-' | sort -n | tail -3
```

Use the next integer, zero-padded to 3 digits.

**2. Create `docs/decisions/NNN-short-slug.md`** in this shape (see
`docs/decisions/025-license-busl.md` for a full real example):

```markdown
# ADR-NNN: <Short decision title>

**Date:** YYYY-MM-DD
**Status:** Accepted

---

## Context

Why this decision was needed. What constraints or prior incident forced it.

## Alternatives considered

| Option | ... relevant comparison columns ... |
|---|---|
| Option A | Ruled out because... |
| Option B | **Chosen** |

(A table works well when comparing options across a few axes; a numbered
list is fine for simpler decisions — match whichever existing ADRs nearby
in topic use.)

## Decision

What was decided, stated clearly and specifically (exact values, exact
scope — not just the general direction).

## Consequences

- What this enables
- What this forecloses
- Known trade-offs accepted
```

**3. If this decision supersedes an earlier ADR**, add a line to the
superseded one noting it (see how ADR-026 supersedes ADR-018 — CLAUDE.md's
billing-provider Don't entry references both). Don't delete the old ADR;
ADRs are a historical log.

**4. Cross-link.** If the decision resolves an item in
`docs/decisions/backlog.md`, remove it from there. If it affects a story in
`docs/stories/`, add a reference line pointing at the new ADR (see how
`ep-19-sms-alerts.md` links `[ADR-029]` and `[ADR-032]` inline).

**5. If the decision changes a rule contributors need to follow**, add it
to `CLAUDE.md`'s Conventions or Don't section too — ADRs record *why*,
CLAUDE.md enforces *what*, and they should agree.
