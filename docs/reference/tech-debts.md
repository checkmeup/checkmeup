---
title: Tech Debt
type: reference
status: active
updated: 2026-07-10
tags: [architecture, maintainability, backend]
---

# Tech debt

Known architecture/code smells that aren't worth an ADR or an immediate fix, but shouldn't be forgotten. Add an entry when you spot something during other work rather than stopping to fix it; remove an entry once it's addressed (reference the commit/PR in the removal, not here — `git log` is the record of what was fixed and when).

---

## Backend (Go)

### Maintainability

- **`checkOverdue`/`buildUptimeDownAlert`/`buildPortDownAlert` each hand-build their own `AlertMessage`** (Telegram/EmailSubject/EmailHTML/Webhook/Slack/SMS, one `fmt.Sprintf` per field) — same shape as `expiredMessages`/`expiringSoonMessages` (`worker.go`, shared by SSL/domain since 2026-07-10) but with genuinely different reason text per type ("missed its ping" vs a check failure reason vs "port unexpectedly open"), not just a field-name swap.
  → Leave as-is for now — forcing these into one shared template would either lose per-type phrasing or need as many parameters as the duplication saves. Revisit if a fourth near-identical case shows up.
