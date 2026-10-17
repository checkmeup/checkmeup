# checkmeup

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/6354d576310e41ab86a9fdd8572366b6)](https://app.codacy.com/gh/checkmeup/checkmeup/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/6354d576310e41ab86a9fdd8572366b6)](https://app.codacy.com/gh/checkmeup/checkmeup/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?logo=conventionalcommits&logoColor=white)](https://conventionalcommits.org)

Cron, uptime, SSL, domain expiry, port (TCP) and DNS record monitors, with Telegram, email, Slack, webhook and SMS alerts, and public status pages for your clients.

## Stack

| Layer | Tech                                             |
| ----- | ------------------------------------------------ |
| API   | Go · Chi · sqlc · goose · PostgreSQL             |
| Web   | Vue 3 · Vite · Pinia · TanStack Query · Tailwind |
| Infra | Hetzner · Kamal · kamal-proxy                    |

## Development

```bash
make setup  # once: tools from mise.toml, bun install, apps/api/.env
make db     # PostgreSQL in Docker on localhost:5432
make dev    # API on :8080 · Web on :5173
make test   # lint + test + coverage
make build  # production build
```

## Documentation

| Doc                                                                | What's there                                          |
| ------------------------------------------------------------------ | ----------------------------------------------------- |
| [docs/INDEX.md](docs/INDEX.md)                                     | Docs navigation cache — folder roles, load order      |
| [docs/roadmap.md](docs/roadmap.md)                                 | Current priorities — Now / Next / Later               |
| [.about/prd/](.about/prd/)                                         | Feature requirements (PRDs)                           |
| [.about/adr/](.about/adr/)                                         | Decisions (ADRs), one file per subject                |
| [docs/reference/design.md](docs/reference/design.md)               | Design tokens — colors, logo usage                    |
| [docs/reference/deploy.md](docs/reference/deploy.md)               | Deployment guide                                      |
| [docs/reference/billing-setup.md](docs/reference/billing-setup.md) | Paddle activation checklist                           |
| [docs/incidents/](docs/incidents/)                                 | Production incident write-ups (one file per incident) |
| [docs/knowledge/mvp-history.md](docs/knowledge/mvp-history.md)     | How the MVP was built (archived, frozen record)       |
| [CLAUDE.md](CLAUDE.md)                                             | Conventions and guardrails for contributors           |

## License

[Business Source License 1.1](LICENSE.md) © 2026 Andrew Molyuk — free for self-hosted/internal use; converts to Apache License 2.0 on 2030-07-01. See [LICENSE.md](LICENSE.md) for the Additional Use Grant (the one restriction: no reselling Checkmeup itself as a competing hosted monitoring service).
