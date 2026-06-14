# ADR-012: Resend for transactional email

**Status:** Accepted  
**Date:** 2026-06-14

## Context

Transactional email is needed for password reset (US-0105). Future use: welcome email. Alerts remain Telegram-only.

Options considered:

- **Zoho SMTP** — already have `@checkmeup.net` mailboxes in Zoho. No new service. Downside: shared IP reputation, poor transactional deliverability, mixes inbox mail with code-triggered sends.
- **Resend** — dedicated transactional API, simple REST, official Go SDK, free tier 3,000 emails/month / 100/day, automatic DKIM/SPF via one DNS record.
- **Postmark** — solid deliverability, but only 100 free emails/month (tight for testing).
- **AWS SES** — cheapest at scale, but extra infrastructure overhead for MVP.

## Decision

Use **Resend** for all transactional email. Zoho remains for human-facing mailboxes (`support@`, `hello@`).

Sending domain: `checkmeup.net` via Resend's DKIM DNS record.  
From address: `checkmeup <noreply@checkmeup.net>`  
Config: `RESEND_API_KEY` environment variable.

## Consequences

- One DNS record (DKIM TXT) to add in the domain registrar.
- `RESEND_API_KEY` required in `.env` and Kamal secrets.
- If the API key is missing in development, email sending is skipped with a log warning (no crash).
- Free tier covers MVP comfortably; upgrade is pay-as-you-go if needed.
