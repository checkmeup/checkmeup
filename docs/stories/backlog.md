# Stories backlog

> Historical: MVP delivery order was EP-02 → EP-03 → EP-04 → EP-05 → EP-06 → EP-07 (EP-01 Auth was a prerequisite for everything and built first). Post-MVP epics (EP-09+) follow the priority order in [roadmap.md](../roadmap.md) instead.

## Epics

| ID                                    | Epic                            | Stories done |
| -------------------------------------- | ------------------------------- | ------------- |
| [EP-01](ep-01-auth.md)                | Authentication & account setup  | 6/6           |
| [EP-02](ep-02-cron-monitor.md)        | Cron monitor                    | 8/8           |
| [EP-03](ep-03-uptime-monitor.md)      | Uptime monitor                  | 6/6           |
| [EP-04](ep-04-ssl-monitor.md)         | SSL monitor                     | 5/5           |
| [EP-05](ep-05-telegram-alerts.md)     | Telegram alerts                 | 4/4           |
| [EP-06](ep-06-status-page.md)         | Status page                     | 5/5           |
| [EP-07](ep-07-billing.md)             | Billing & plan limits           | 3/3           |
| [EP-08](ep-08-security-hardening.md)  | Security hardening              | 4/4           |
| [EP-09](ep-09-maintenance-windows.md) | Maintenance windows             | 5/5           |
| [EP-10](ep-10-theme.md)               | Light & dark theme              | 3/3*          |
| [EP-11](ep-11-keyword-monitoring.md)  | Keyword monitoring              | 5/5           |
| [EP-12](ep-12-team-management.md)     | Team management                 | 0/5           |
| [EP-13](ep-13-email-alerts.md)        | Email alerts                    | 5/5           |
| [EP-14](ep-14-webhook-alerts.md)      | Webhook alerts                  | 5/5           |
| [EP-15](ep-15-whatsapp-alerts.md)     | WhatsApp alerts                 | 0/5           |
| [EP-16](ep-16-signal-alerts.md)       | Signal alerts                   | 0/5           |
| [EP-17](ep-17-slack-alerts.md)        | Slack alerts                    | 5/5           |
| [EP-18](ep-18-teams-alerts.md)        | Microsoft Teams alerts          | 0/5           |
| [EP-19](ep-19-sms-alerts.md)          | SMS alerts                      | 7/8****       |
| [EP-20](ep-20-viber-alerts.md)        | Viber alerts                    | 0/5           |
| [EP-21](ep-21-terms-and-privacy.md)   | Terms and Privacy               | 4/4           |
| [EP-22](ep-22-faq.md)                 | FAQ                             | 3/3           |
| [EP-23](ep-23-suggest-a-feature.md)   | Suggest a feature               | 3/3           |
| [EP-24](ep-24-incident-management.md) | Incident management             | 5/5           |
| [EP-25](ep-25-two-factor-auth.md)     | Two-factor authentication       | 0/5           |
| [EP-26](ep-26-public-api-keys.md)     | Public API and API keys         | 3/5*****      |
| [EP-27](ep-27-annual-billing.md)      | Annual billing                  | 3/3           |
| [EP-28](ep-28-notification-channels.md) | Notification channels         | 4/5***        |
| [EP-29](ep-29-domain-expiry-monitoring.md) | Domain expiry monitoring   | 3/5**         |
| [EP-30](ep-30-status-badges.md)       | Public status badges            | 4/4           |
| [EP-31](ep-31-assertion-checks.md)    | Assertion-based API checks      | 4/5           |
| [EP-32](ep-32-multi-region-checking.md) | Multi-region checking         | 0/4           |
| [EP-33](ep-33-port-monitoring.md)     | Port (TCP) monitoring           | 4/5******     |
| [EP-34](ep-34-zombie-job-detection.md) | Zombie (stuck) job detection   | 0/4           |
| [EP-35](ep-35-overlap-detection.md)   | Overlap detection               | 0/3           |
| [EP-36](ep-36-blog-prerendering.md)   | Prerender public routes for crawlers/social previews | 1/1 |
| [EP-37](ep-37-configurable-uptime-checks.md) | Configurable uptime check parameters     | 3/3 |

`x/y` = stories done vs. total stories in the epic (a story counts as done once every acceptance criterion is checked). \* EP-10: US-1002's "public status page" criterion was redefined, not deferred — see the epic file for why. \*\* EP-29: US-2902 and US-2903 each shipped with one AC intentionally not built (WHOIS fallback, registry-hold-status alert) — see the epic file's shipped note. \*\*\* EP-28: US-2804's legacy-column-drop AC is deliberately deferred to a later migration, after the cutover has proven stable in production — see the epic file. \*\*\*\* EP-19: US-1906 shipped with its destination-weighted cost-band AC not built (flat 1-credit-per-send instead — needs a hand-built per-country pricing table, deferred as a data-entry task) — see the epic file's implementation note. \*\*\*\*\* EP-26: US-2604 shipped with its "scope" column and `••••`-masked key format not built as specified (no scope column, since key scoping (US-2603) isn't built; key is shown as a bare prefix instead) — see the epic file. \*\*\*\*\*\* EP-33: US-3302 shipped with its DNS-resolution-failure-as-distinct-error-state AC not built (treated the same as connection-refused/timeout) — see the epic file's shipped note.

See [roadmap.md](../roadmap.md) for the Now / Next / Later priority order ([ADR-022](../decisions/022-post-mvp-docs-organization.md)).
