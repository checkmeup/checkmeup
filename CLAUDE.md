# checkmeup.net — developer monitoring platform

Cron, uptime, and SSL monitors with execution logs, Telegram alerts, and white-label status pages for agencies.

**MVP order:** Cron monitor → Uptime monitor → SSL expiry monitor → Status page  
**Pricing:** Hobbyist $0 / Indie $12 / Studio $39 / Agency $99

---

## Commands

```bash
make dev      # start all apps (hot reload)
make test     # lint + test + merge coverage
make lint     # lint only
make build    # production build
make clean    # wipe node_modules, dist, coverage
```

---

## Stack

**Backend (`apps/api`):** Go · Chi · sqlc · goose · PostgreSQL · JWT auth · Telegram alerts · air (hot reload)  
**Frontend (`apps/web`):** Vue 3 · Vite · Pinia · TanStack Query · Radix Vue · Tailwind  
**Infra:** Hetzner CX23 · Kamal · Traefik  
**Test tooling:** golangci-lint · gcov2lcov (Go coverage → lcov) · Vitest

> `make test` requires PostgreSQL running (`docker-compose up db` or inside the devcontainer).

Full rationale in [`docs/decisions/`](docs/decisions/). Open questions in [`docs/decisions/backlog.md`](docs/decisions/backlog.md).

---

## Conventions

- **Commits:** conventional commits enforced by commitlint — `feat:`, `fix:`, `chore:`, etc.
- **JS/TS lint+format:** oxlint + oxfmt (run automatically on commit via lint-staged)
- **Go lint:** golangci-lint
- **Multi-tenancy:** every tenant-scoped query **must** filter by `org_id`
- **Colors/logo:** never hardcode hex values — use tokens from [`docs/design.md`](docs/design.md)

---

## Code quality (Codacy)

CI uploads coverage to Codacy after every push. `CODACY_API_TOKEN` (account-level) is in `apps/api/.env`.

**Fetch current issues before starting a fix session:**

```bash
source apps/api/.env
curl -s -X POST \
  -H "api-token: $CODACY_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filters":{"categories":[],"levels":[],"languages":[]}}' \
  "https://app.codacy.com/api/v3/analysis/organizations/gh/checkmeup/repositories/checkmeup/issues/search?limit=100" \
  | python3 -c "
import json, sys
issues = json.load(sys.stdin)['data']
priority = [i for i in issues if i['patternInfo']['level'] in ('Error','High','Warning')]
for i in sorted(priority, key=lambda x: x['patternInfo']['level']):
    print(f\"[{i['toolInfo']['name']}][{i['patternInfo']['level']}] {i['filePath']}:{i['lineNumber']}: {i['message'][:100]}\")
"
```

**Triage guide:**
- **TSQLLint** — always ignore, SQL Server rules applied to PostgreSQL migrations
- **Opengrep cookies in `*_test.go`** — ignore, synthetic request cookies in tests intentionally lack HttpOnly/Secure
- **Trivy on `go.mod`** — real; upgrade the flagged dependency or pin a patched version
- **Opengrep/ESLint in production code** — investigate before dismissing

---

## Don't

- Add Redis, a job queue, or any external broker — goroutine workers are intentional ([ADR-001](docs/decisions/001-worker-model.md))
- Use an ORM — sqlc only ([ADR-004](docs/decisions/004-sqlc-over-orm.md))
- Skip `org_id` filters in DB queries — silent data leak across tenants
- Add subdomains for status pages — `/status/:slug` path is intentional ([ADR-005](docs/decisions/005-status-page-same-domain.md))
- Use `Authorization` header for auth — the `access_token` httpOnly cookie is the only auth mechanism ([ADR-003](docs/decisions/003-auth-jwt-httponly-cookie.md))
