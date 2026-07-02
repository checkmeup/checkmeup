#!/usr/bin/env bash
# Deletes old ghcr.io/checkmeup/checkmeup container image versions, keeping
# only the N most recent (default 5). Kamal never prunes GHCR itself — every
# deploy leaves the previous image behind (docs/deploy.md).
#
# Requires a classic GitHub PAT with read:packages + delete:packages scopes,
# passed via GHCR_CLEANUP_TOKEN or (reusing the existing deploy secret)
# KAMAL_REGISTRY_PASSWORD — the latter needs delete:packages added to its
# scopes on GitHub for this to work, since write:packages alone can't delete.
set -euo pipefail

ORG="checkmeup"
PACKAGE="checkmeup"
KEEP="${1:-5}"
TOKEN="${GHCR_CLEANUP_TOKEN:-${KAMAL_REGISTRY_PASSWORD:-}}"

if [ -z "$TOKEN" ]; then
  echo "ghcr-clean: no token found — set GHCR_CLEANUP_TOKEN or KAMAL_REGISTRY_PASSWORD" >&2
  exit 1
fi

api() {
  curl -sf -H "Authorization: Bearer $TOKEN" -H "Accept: application/vnd.github+json" "$@"
}

to_delete=$(api "https://api.github.com/orgs/$ORG/packages/container/$PACKAGE/versions?per_page=100" | python3 -c "
import json, sys
versions = json.load(sys.stdin)
versions.sort(key=lambda v: v['created_at'], reverse=True)
for v in versions[$KEEP:]:
    print(v['id'])
")

if [ -z "$to_delete" ]; then
  echo "ghcr-clean: nothing to prune (≤$KEEP versions exist)"
  exit 0
fi

for id in $to_delete; do
  echo "ghcr-clean: deleting version $id"
  api -X DELETE "https://api.github.com/orgs/$ORG/packages/container/$PACKAGE/versions/$id" > /dev/null
done

echo "ghcr-clean: done, kept the $KEEP most recent versions"
