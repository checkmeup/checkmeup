#!/usr/bin/env python3
"""Flag sqlc queries on org-scoped tables that look up by id without org_id."""
# CLAUDE.md: "every tenant-scoped query must filter by org_id — silent
# data leak across tenants" — the single most emphasized invariant in
# this repo. Org-scoped tables (those with an org_id column) are found
# by parsing apps/api/migrations/*.sql directly, not hardcoded, so a new
# migration adding org_id to a table is picked up automatically.
#
# A query only needs auditing if it does a primary-key lookup (bare
# "id = $N", not "monitor_id"/"user_id" etc — the \b word boundary
# already excludes those) against an org-scoped table. Findings not on
# the KNOWN_EXCEPTIONS list are real: either add the org_id filter, or
# — if the id is verifiably not attacker-controlled — add it to
# KNOWN_EXCEPTIONS with which category applies and why (see SKILL.md).
import glob
import re
import sys

TABLE_RE = re.compile(r"CREATE TABLE (?:IF NOT EXISTS )?(\w+)\s*\((.*?)\n\);", re.S)
PK_LOOKUP_RE = re.compile(r"\bid\s*=\s*\$\d")

# query name -> one-line rationale. Verified against actual call sites
# when this audit was authored — see SKILL.md's "Known exceptions".
KNOWN_EXCEPTIONS = {
    "TouchAPIKeyLastUsed": "id is the just-authenticated key's own ID (apikey.go RequireAPIKey), not caller input",
    "GetUserByID": "id is the authenticated user's own ID from JWT claims, not a caller-supplied arbitrary user",
    "UpdateUserPassword": "id is the authenticated user's own ID from JWT claims",
    "AcceptUserTerms": "id is the authenticated user's own ID from JWT claims",
    "UpdateCronMonitorDown": "id sourced from ListOverdueCronMonitors, worker-internal, never request input",
    "IncrementCronConsecutiveFailures": "id sourced from an already-fetched worker-internal monitor row",
    "UpdateDomainMonitorCheck": "id sourced from an already-fetched worker-internal monitor row",
    "RecordPortCheckUp": "id sourced from an already-fetched worker-internal monitor row",
    "RecordPortCheckFailure": "id sourced from an already-fetched worker-internal monitor row",
    "MarkPortMonitorDown": "id sourced from an already-fetched worker-internal monitor row",
    "UpdateSSLMonitorCheck": "id sourced from an already-fetched worker-internal monitor row",
    "RecordUptimeCheckUp": "id sourced from an already-fetched worker-internal monitor row",
    "RecordUptimeCheckFailure": "id sourced from an already-fetched worker-internal monitor row",
    "MarkUptimeMonitorDown": "id sourced from an already-fetched worker-internal monitor row",
    "GetUptimeMonitorPublic": "id sourced from a prior status_page_monitors-joined query (status_public.go), not raw request input",
    "GetSSLMonitorPublic": "id sourced from a prior status_page_monitors-joined query, not raw request input",
    "GetDomainMonitorPublic": "id sourced from a prior status_page_monitors-joined query, not raw request input",
    "GetPortMonitorPublic": "id sourced from a prior status_page_monitors-joined query, not raw request input",
}


def org_scoped_tables(migrations_glob: str) -> set:
    tables = set()
    for path in glob.glob(migrations_glob):
        up = open(path).read().split("-- +goose Down")[0]
        for m in TABLE_RE.finditer(up):
            name, body = m.group(1), m.group(2)
            if re.search(r"\borg_id\b", body):
                tables.add(name)
    return tables


def query_blocks(queries_glob: str):
    for path in sorted(glob.glob(queries_glob)):
        text = open(path).read()
        for block in re.split(r"(?=-- name: )", text):
            if block.startswith("-- name: "):
                name = block.splitlines()[0].split()[2]
                yield path, name, block


def audit(migrations_glob: str, queries_glob: str):
    org_tables = org_scoped_tables(migrations_glob)
    findings, exceptions = [], []
    for path, name, block in query_blocks(queries_glob):
        table = next((t for t in org_tables if re.search(rf"\b{t}\b", block)), None)
        if not table or not PK_LOOKUP_RE.search(block) or "org_id" in block:
            continue
        entry = (path, name, table)
        (exceptions if name in KNOWN_EXCEPTIONS else findings).append(entry)
    return findings, exceptions


def report(findings, exceptions) -> int:
    if exceptions:
        print("Known exceptions (not flagged as failures):")
        for path, name, table in exceptions:
            print(f"  {path}: {name} [{table}] — {KNOWN_EXCEPTIONS[name]}")
        print()

    if findings:
        print("Queries on org-scoped tables with NO org_id filter:")
        for path, name, table in findings:
            print(f"  {path}: {name} [{table}]")
        print(f"\n{len(findings)} unaudited/unfiltered quer{'y' if len(findings) == 1 else 'ies'}.")
        return 1

    print("All primary-key lookups on org-scoped tables are org_id-filtered or a known exception.")
    return 0


if __name__ == "__main__":
    migrations = sys.argv[1] if len(sys.argv) > 1 else "apps/api/migrations/*.sql"
    queries = sys.argv[2] if len(sys.argv) > 2 else "apps/api/queries/*.sql"
    sys.exit(report(*audit(migrations, queries)))
