# MVP backlog

**MVP delivery order:** EP-02 → EP-03 → EP-04 → EP-05 → EP-06 → EP-07  
(EP-01 Auth is a prerequisite for everything and is built first.)

## Epics

| ID                                | Epic                           | Stories |
| --------------------------------- | ------------------------------ | ------- |
| [EP-01](ep-01-auth.md)            | Authentication & account setup | 6       |
| [EP-02](ep-02-cron-monitor.md)    | Cron monitor                   | 8       |
| [EP-03](ep-03-uptime-monitor.md)  | Uptime monitor                 | 6       |
| [EP-04](ep-04-ssl-monitor.md)     | SSL monitor                    | 5       |
| [EP-05](ep-05-telegram-alerts.md) | Telegram alerts                | 4       |
| [EP-06](ep-06-status-page.md)     | Status page                    | 5       |
| [EP-07](ep-07-billing.md)         | Billing & plan limits          | 3       |
| [EP-08](ep-08-security-hardening.md) | Security hardening          | 4       |

## Story status key

| Symbol | Meaning     |
| ------ | ----------- |
| ` `    | To do       |
| `~`    | In progress |
| `x`    | Done        |

## Plan limits reference

(Revised Jun 2026 after competitor review — see [ADR-019](../decisions/019-plan-limits.md))

|                    | Hobbyist | Indie | Studio | Agency    |
| ------------------ | -------- | ----- | ------ | --------- |
| Price              | $0       | $9/mo | $29/mo | $79/mo    |
| Monitors           | 10       | 30    | 100    | unlimited |
| Min check interval | 5 min    | 1 min | 1 min  | 1 min     |
| Status pages       | 1        | 3     | 10     | unlimited |
