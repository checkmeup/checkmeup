#!/usr/bin/env bash
# Run Stryker against apps/web, working around this environment's quirks.
# Any arguments are passed straight through to stryker (e.g. --mutate <glob>).
set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DIR="$(cd "$SKILL_DIR/../../../apps/web" && pwd)"
cd "$WEB_DIR"

# `npx stryker` resolves to an abandoned 2019 package of that name on the
# public registry, not to the local install. Always call the binary directly.
STRYKER="./node_modules/.bin/stryker"
if [ ! -x "$STRYKER" ]; then
  echo "Stryker not installed. Run: cd apps/web && bun add -d @stryker-mutator/core @stryker-mutator/vitest-runner" >&2
  exit 1
fi

# Stryker shells out to `ps` for process cleanup; the devcontainer has no
# procps and no sudo to install it. Put a /proc-reading stand-in on PATH.
if ! command -v ps >/dev/null 2>&1; then
  SHIM_BIN="$(mktemp -d)/bin"
  mkdir -p "$SHIM_BIN"
  cp "$SKILL_DIR/ps_shim.py" "$SHIM_BIN/ps"
  chmod +x "$SHIM_BIN/ps"
  export PATH="$SHIM_BIN:$PATH"
  echo "note: using ps shim from $SHIM_BIN (no procps in this container)" >&2
fi

# A sandbox left behind by an interrupted run gets copied into the next run's
# sandboxes, which breaks the dry run and inflates the suite. Always start clean.
rm -rf .stryker-tmp

"$STRYKER" run "$@"
status=$?

rm -rf .stryker-tmp
exit $status
