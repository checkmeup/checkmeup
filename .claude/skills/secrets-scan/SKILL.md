---
name: secrets-scan
description: Scan staged changes (or the full git-tracked tree) for credential-shaped strings and accidentally-tracked secret files (.env, .pem, .key, id_rsa). Use when asked to "check for secrets", "scan for leaked credentials", or before pushing a branch — good pre-flight before the pr-merge skill.
---

# Secrets scan

Scans **git-tracked/staged content only** — never the working tree's
actual `.env` files — for two things: filenames that should never be
committed, and credential-shaped strings that got typed or pasted into
tracked content.

A `PreToolUse` hook (`.claude/hooks/pre_push_guard.py`) already runs
`scan.py tree` before every `git push` and blocks the push on a hit, so
the "good pre-flight before pr-merge" step below now happens automatically
rather than depending on remembering to run it. This skill is still useful
on demand: `staged` mode (the hook only covers `tree`), or a full audit
outside of a push.

## Steps

**1. Before committing/pushing, scan staged changes:**

```bash
python3 .claude/skills/secrets-scan/scan.py staged
```

**2. Periodically (or when asked to audit the whole repo), scan
everything already tracked:**

```bash
python3 .claude/skills/secrets-scan/scan.py tree
```

Exit code `0` + "Clean" means nothing found. Exit code `1` lists findings
under two headers:

- **Forbidden files** — `.env`/`.env.*` (excluding `.env.example`),
  `.pem`, `.key`, `id_rsa`, `id_ed25519` staged or tracked. These should
  never be committed regardless of content — remove from staging
  (`git restore --staged <file>`) or, if already merged to history,
  treat as a real incident (see below).
- **Credential-shaped strings** — Twilio Account/API-Key SIDs, GitHub
  tokens, AWS access keys, Resend API keys, PEM private-key blocks,
  Slack webhook URLs, DB URLs with embedded credentials, Telegram bot
  tokens. Rotate the credential and remove it from the diff/file.

**3. If a real secret is found already merged to `main`** (not just
staged), stop and tell the user directly — rewriting merged history
(`git filter-repo`, force-push) is a destructive, shared-history
operation this skill should never do on its own. The credential should
also be rotated at the provider regardless of whether history gets
rewritten, since it's already been exposed.

## What this deliberately does NOT flag

This repo has one documented, non-secret local-dev Postgres credential
(`checkmeup:checkmeup`, matching `.devcontainer/docker-compose.yml`) that
appears by design in `apps/api/.env.example`, CI config, and test
fixtures — exempted everywhere, same reasoning as CLAUDE.md's "ignore
cookies in `*_test.go`": a known synthetic value, not a leak. Similarly,
DB-URL and Slack-webhook patterns are skipped in `*_test.go`,
`*.test.tsx?`, `*.md`, and `*.example` files generally, since those are
expected to contain realistic-looking fixtures/placeholders (e.g.
`docs/reference/deploy.md`'s `postgres://checkmeup:<password>@...`).

Provider-token *shapes* (Twilio/GitHub/AWS/private-key/Telegram) have no
such legitimate fixture use anywhere in this repo — verified empty when
this scan was authored — so those stay strict even in tests and docs. If
a new provider integration adds a test fixture that trips one of these,
prefer an obviously-fake value in the fixture (e.g. all-zeros) over
widening the exemption.

## Local verification of scan.py itself

If you edit `scan.py`, check it against the same tools Codacy runs in CI
before pushing — `.devcontainer/Dockerfile` installs `bandit` via `pipx`
(rebuild the devcontainer to pick up a Dockerfile change):

```bash
bandit .claude/skills/secrets-scan/scan.py
```

Any `subprocess` call needs a `# nosec` with a one-line rationale (see
the existing ones in `scan.py`) rather than a silent suppression —
Bandit's subprocess warnings are worth reading once per call site, not
blanket-ignoring.

## Why no off-the-shelf tool

`gitleaks`/`trufflehog`/`detect-secrets` aren't installed in this
environment, and a generic "any assigned high-entropy string" heuristic
would flag `apps/api/.env.example`'s placeholder values (`change-me-in-
production-use-32-plus-random-chars`, `your-codacy-api-token`)
constantly — noisy enough that it trains people to ignore the tool. This
scan uses specific, known credential *shapes* instead. If one of the
listed tools becomes available, prefer it for defense in depth, but keep
this scan too — it's tuned to this repo's actual false-positive sources
in a way a generic tool wouldn't be out of the box.
