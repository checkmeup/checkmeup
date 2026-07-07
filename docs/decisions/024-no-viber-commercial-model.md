# ADR-024 — Do not integrate Viber alerts

**Status:** Accepted  
**Date:** 2026-06-24

## Context

EP-20 started a Viber notification channel using the Viber Bot API. During implementation it was discovered that on **5 February 2024**, Viber switched Viber bots to a commercial-only model: new bots can only be created under commercial terms, requiring a business agreement with Viber/Rakuten.

Reference: <https://help.viber.com/hc/en-us/articles/15247629658525-Bot-commercial-model>

This means there is no self-serve path for checkmeup to obtain a bot token, and we cannot offer the feature to users without a commercial Viber partnership.

## Decision

Remove all Viber integration work (EP-20) from the codebase. The `viber` package, its migration, the user story, and all references in handlers, worker, config, frontend, and docs are deleted.

## Consequences

- Viber is not offered as a notification channel.
- No `VIBER_BOT_TOKEN` env var or `/webhook/viber` route exists.
- If Viber reintroduces self-serve bots in the future, EP-20 can be re-implemented from git history.
