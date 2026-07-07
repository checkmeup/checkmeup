---
name: new-story
description: Scaffold a new epic/story file under docs/stories/ from the repo's template, numbered sequentially, and register it in docs/stories/backlog.md. Use when asked to "create a new epic", "add a story for X", "write a story", or "add EP-NN".
---

# New epic / story

Epics live one-per-file at `docs/stories/ep-NN-slug.md`, each containing
one or more `US-NNNN` user stories. Real files (like
`ep-19-sms-alerts.md`) follow `_template.md`'s structure fairly closely,
unlike ADRs — use the template directly here.

## Steps

**1. Find the next epic number.**

```bash
ls docs/stories | grep -E '^ep-[0-9]+-' | sort -t- -k2 -n | tail -3
```

**2. Find the next available `US-NNNN` story-ID block.** Story IDs are
4-digit and roughly track epic number (`EP-19` uses `US-1901`–`US-1908`) —
check the most recent epics for the current numbering convention rather
than assuming a fixed scheme.

**3. Create `docs/stories/ep-NN-slug.md`** from `docs/stories/_template.md`:

```markdown
# EP-NN: <Epic Title>

One-paragraph context: what problem this solves and why now. Link any
ADRs or prior epics this builds on, inline — e.g. "builds on the
multi-channel model in [EP-28](ep-28-notification-channels.md)".

---

### US-NNNN: <Story Title>

**As a** <role>, **I want** <goal> **so that** <reason>.

**Estimate:** N h

**Acceptance criteria:**

- [ ] AC 1
- [ ] AC 2
```

Repeat the `US-NNNN` block per story. Link to the deciding ADR inline
where a design choice needs one (see `ep-19-sms-alerts.md`'s "Provider and
opt-in flow decided in [ADR-029](../decisions/029-sms-alerts-twilio.md)"
pattern) rather than re-explaining the decision in the story itself.

**4. Register in `docs/stories/backlog.md`**, adding a row to the Epics
table:

```text
| [EP-NN](ep-NN-slug.md) | <Epic name> | 0/N |
```

Keep the table's column alignment consistent with neighboring rows.

**5. Placement relative to `docs/roadmap.md`.** Post-MVP epics (EP-09+)
are prioritized via the Now/Next/Later structure in `docs/roadmap.md`
rather than a fixed delivery order — if this is a real near-term
priority, add it there too, not just to the backlog table.

**6. When a story ships**, update its epic file with a "**Shipped
YYYY-MM-DD.**" note (see `ep-19-sms-alerts.md`'s pattern — including
which ACs shipped as originally scoped vs. which were simplified and
why), and update the backlog table's `stories done` fraction.
