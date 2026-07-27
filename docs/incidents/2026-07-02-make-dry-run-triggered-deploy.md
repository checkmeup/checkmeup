# 2026-07-02: `make -n deploy` triggered a real deploy

**Impact window:** momentary — one unintended deploy to production, no downtime
**Detected by:** self (noticed the deploy output while attempting a dry run)

## What broke

`make -n deploy`, run to preview the `deploy` target without executing it,
actually deployed to production (`checkmeup.net`) via Kamal.

## Root cause

GNU Make always executes any recipe line that references `$(MAKE)` or
`${MAKE}`, even under `-n` (dry-run) — this lets a recursive sub-make show
what it *would* print. The `deploy` target's recipe references `$(MAKE)`
for its post-deploy `ghcr-clean` step, so `-n` didn't skip the real deploy
command it was supposed to only preview.

## Follow-up

- `CLAUDE.md`'s Don't section got a first-line warning against running
  `make deploy`/`make ghcr-clean`/`kamal <anything>`/`docker build`\`buildx`
  against `config/deploy.yml` at all outside the human operator — `-n` is
  explicitly called out as not a safe way to preview this target.
- 2026-07-27: turned into a deterministic guard — a `PreToolUse` hook
  (`.claude/hooks/block_make_deploy.py`, `block_kamal.py`,
  `block_docker_deploy_build.py`) now blocks all of these commands for
  Claude at the tool-call level, including flag-shuffled dry runs like
  `make -n deploy`, rather than relying on the written warning alone.
