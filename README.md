# checkmeup

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/6354d576310e41ab86a9fdd8572366b6)](https://app.codacy.com/gh/checkmeup/checkmeup/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/6354d576310e41ab86a9fdd8572366b6)](https://app.codacy.com/gh/checkmeup/checkmeup/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?logo=conventionalcommits&logoColor=white)](https://conventionalcommits.org)

Cron, uptime, SSL, domain expiry, and port (TCP) monitors with Telegram alerts and white-label status pages.

## Stack

| Layer | Tech                                             |
| ----- | ------------------------------------------------ |
| API   | Go · Chi · sqlc · goose · PostgreSQL             |
| Web   | Vue 3 · Vite · Pinia · TanStack Query · Tailwind |
| Infra | Hetzner · Kamal · Traefik                        |

## Development

```bash
# Prerequisites: Docker (for PostgreSQL)
docker-compose up -d db

make dev    # API on :8080 · Web on :5173
make test   # lint + test + coverage
make build  # production build
```

## Documentation

| Doc                                                | What's there                                          |
| --------------------------------------------------- | ------------------------------------------------------ |
| [docs/roadmap.md](docs/roadmap.md)                 | Current priorities — Now / Next / Later               |
| [docs/stories/backlog.md](docs/stories/backlog.md) | Epics and user stories                                |
| [docs/decisions/](docs/decisions/)                 | Architecture decision records (ADRs)                  |
| [docs/design.md](docs/design.md)                   | Design tokens — colors, logo usage                     |
| [docs/deploy.md](docs/deploy.md)                   | Deployment guide                                       |
| [docs/billing-setup.md](docs/billing-setup.md)     | LemonSqueezy activation checklist                       |
| [docs/reports/](docs/reports/)                     | Monthly snapshots — what shipped, ADRs added            |
| [docs/incidents/](docs/incidents/)                 | Production incident write-ups (one file per incident)  |
| [docs/mvp-history.md](docs/mvp-history.md)         | How the MVP was built (archived, frozen record)       |
| [CLAUDE.md](CLAUDE.md)                             | Conventions and guardrails for contributors            |

## License

MIT © 2026 Andrew Molyuk
