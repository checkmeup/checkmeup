# Public API with API keys

## Problem

Scripts, CI pipelines and other tools need monitor status without a browser session.

## Users

Developers integrating Checkmeup into their own tooling; third parties acting for an org.

## Success signal

A user creates a key and reads a monitor's status with one `curl` from the docs.

## Scope

- API keys per org: created with a label, shown once, stored hashed, prefixed `cmu_live_`, revocable (ADR auth).
- Read-only `GET` endpoints under `/api/v1/public` for monitor status, authenticated by `X-API-Key`.
- Per-key rate limits.
- A docs section listing the endpoints with examples.

## Non-goals

- Write endpoints, and the read-only or read-write key scope that lands with the first of them (ADR auth).
- OAuth or per-endpoint permissions.

## Acceptance criteria

- Creating a key shows its full value once, starting with `cmu_live_`; afterwards only its prefix is shown.
  → test: apps/api/internal/handler/api_keys_test.go "TestCreateAPIKey_ReturnsRawKeyOnceAndPrefix"
  → test: apps/web/src/components/ApiKeysCard.test.ts "creates a key and shows the raw value once"
  → test: apps/web/src/components/ApiKeysCard.test.ts "renders key rows with label and masked prefix"
- A request without a key, or with an invalid or revoked one, gets `401`.
  → test: apps/api/internal/middleware/apikey_test.go "TestRequireAPIKey_NoHeader"
  → test: apps/api/internal/middleware/apikey_test.go "TestRequireAPIKey_InvalidKey"
  → test: apps/api/internal/middleware/apikey_test.go "TestRequireAPIKey_RevokedKey"
- A valid key reads its own org's monitors only; another org's monitor is not found.
  → test: apps/api/internal/middleware/apikey_test.go "TestRequireAPIKey_ValidKey"
  → test: apps/api/internal/handler/public_status_test.go "TestGetCronStatus_ReturnsStatusAndPingMetadata"
  → test: apps/api/internal/handler/public_status_test.go "TestGetCronStatus_WrongOrgNotFound"
- Revoking a key removes it from the list and stops it working on the next request; another org can't revoke it.
  → test: apps/api/internal/handler/api_keys_test.go "TestListAPIKeys_ExcludesRevoked"
  → test: apps/api/internal/handler/api_keys_test.go "TestRevokeAPIKey_CrossOrgIsolation"
- A 101st active key in one org is refused (ADR limits).
  → test: apps/api/internal/handler/api_keys_test.go "TestCreateAPIKey_RejectsA101stActiveKey"
- The docs show the status endpoint with a `curl` example using `X-API-Key`.
  → test: apps/web/src/views/DocsView.test.ts "documents the public API status endpoint with a curl example"
- A 61st request in a minute on one key gets `429` with `Retry-After`.
  → test: apps/api/internal/server/server_test.go "TestRoutes_PublicAPIIsRateLimitedPerKey/a 61st request in a minute on one key gets 429 with Retry-After"
- The key list shows each key's label, created and last-used dates; recording last use never slows the request.
  → test: apps/web/src/components/ApiKeysCard.test.ts "shows each key's label, created date and last-used time"
  → test: apps/api/internal/middleware/apikey_test.go "TestRequireAPIKey_RecordingLastUseNeverSlowsTheRequest"
