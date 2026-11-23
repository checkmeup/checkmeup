# Show a domain's lifecycle stage in the domain expiry checker

2026-11-23

## Symptom

The domain expiry checker (`/tools/domain-expiry-checker`, PRD free-tools) shows a domain's expiry date, days left and registrar, but not whether a lapsed domain is still in its grace period, in redemption or about to be deleted, which is what people searching "is my domain expired" want to know. Competitors in the top 10 for those searches (Pulsetic, Domainyze, UptimeObserver) show it. The checker should show the domain's lifecycle stage, and its FAQ should point to that instead of to a WHOIS record.

## Evidence

- Search Console, 28 days to 2026-11-21: about 170 of the checker's ~1300 impressions are for "is my domain expired", "check if domain has expired" and the like; its one click that week came from "domain redemption period check".
- `apps/api/internal/rdap/rdap.go` reads only `events` and `entities` from the RDAP answer and drops `status` (and `nameservers`).
- `apps/api/internal/whois/` reads no `Domain Status:` lines for the TLDs looked up over WHOIS (ADR domain-lookups).
- `apps/web/src/content/domainExpiryFaqs.ts` answers questions about expiry and redemption by sending the reader to a WHOIS record.

## Acceptance criteria

- A domain whose RDAP answer lists the status `redemption period` shows the stage "Redemption" on the checker, beside its expiry date, days left and registrar.
- Likewise `pending delete` shows "Pending delete", `auto renew period` or `renew period` shows "Grace period", and an answer with none of those (`active`, `client transfer prohibited` and the like) shows "Active".
- A domain whose answer lists several of those shows the latest stage: pending delete before redemption before grace.
- A domain whose RDAP answer has no `status` shows no stage, and the rest of the result as today.
- A domain under a WHOIS TLD shows its stage only where that registry's real answer, kept as a test fixture, carries `Domain Status:` lines (ADR domain-lookups: no parsing of formats nobody has looked at); otherwise no stage.
- The FAQ answers about whether a domain has expired and about the redemption period say that the checker shows the stage, on the page and in its FAQPage JSON-LD (PRD free-tools).

## Comments

### 2026-11-23 · cause

The RDAP client was written for the domain monitor, which needs only the expiry date and registrar, so it never kept `status`; the checker reuses the same lookup (`worker.LookupDomain`).

### 2026-11-23 · fix option

Keep the EPP status values the RDAP answer lists (RFC 8056 maps them to RDAP's lower-case phrases), return them through `LookupDomain`, and map them to a stage in one place, used by the checker's API response and the page. The domain monitor can ignore them for now: showing the stage on a monitor, or alerting on redemption, is a separate decision.

### 2026-11-23 · scope

Nameservers, also dropped today and shown by competitors, are left out: no search query in the evidence asks for them. Adding them is a separate Ticket if the data asks for it.
