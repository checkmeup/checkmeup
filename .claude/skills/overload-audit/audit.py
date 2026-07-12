#!/usr/bin/env python3
"""Re-verify the concrete, greppable claims in docs/reference/limits.md."""
# Mirrors that doc's "Things that are fine" checklist item-by-item so it
# can't silently drift stale again — it already did once (the doc's own
# note: findings were fixed in code without the doc being updated). Only
# checks claims with a mechanical, unambiguous signal; vague ones ("reads
# are paginated") or structural guarantees ("sqlc is parameterized by
# construction") aren't re-verified here — see SKILL.md's Scope section.
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]  # repo root, from .claude/skills/overload-audit/audit.py
API = ROOT / "apps" / "api"

CHECK_LOOP_FILES = [
    "worker_cron.go",
    "worker_uptime.go",
    "worker_ssl.go",
    "worker_domain.go",
    "worker_port.go",
]

INCIDENT_QUERIES = {
    "queries/monitors.sql": "cron_incidents",
    "queries/uptime.sql": "uptime_incidents",
    "queries/port.sql": "port_incidents",
}

PRUNE_QUERIES = ["DeleteOldCronPings", "DeleteOldUptimeChecks", "DeleteOldPortChecks", "DeleteOldStatusPageIncidents"]

STATUS_PAGE_INCIDENT_QUERIES = ["ListStatusPageIncidents", "ListActiveStatusPageIncidentsForPage"]

# (query file, [:many query names expected to carry a LIMIT 200]) — the
# three added 2026-07-12 alongside FLAT_CAPS below, same audit pass.
NAMED_LIMIT_200_QUERIES = [
    ("queries/incidents.sql", ["ListStatusPageIncidentUpdates"]),
    ("queries/maintenance.sql", ["ListMaintenanceWindows"]),
    ("queries/api_keys.sql", ["ListAPIKeys"]),
]


def check_bounded_concurrency(findings):
    missing = []
    for name in CHECK_LOOP_FILES:
        text = (API / "internal" / "worker" / name).read_text()
        if "make(chan struct{}, checkConcurrency)" not in text:
            missing.append(name)
    if missing:
        findings.append("unbounded check loop(s) — no checkConcurrency semaphore found in: " + ", ".join(missing))


def check_incident_limits(findings):
    for rel, table in INCIDENT_QUERIES.items():
        text = (API / rel).read_text()
        pattern = re.compile(rf"FROM {table}.*LIMIT 200", re.IGNORECASE)
        if not pattern.search(text):
            findings.append(f"{rel}: no LIMIT-200-capped list query found for {table}")


def check_status_page_incident_limits(findings):
    text = (API / "queries" / "incidents.sql").read_text()
    # Each named query's body runs from its "-- name: X :many" header to the
    # next "-- name:" (or EOF) — split on that header so a LIMIT 200 in one
    # query can't be mistaken for coverage of a different one further up.
    blocks = dict(re.findall(r"-- name: (\w+) :\w+\n(.*?)(?=\n-- name:|\Z)", text, re.DOTALL))
    for name in STATUS_PAGE_INCIDENT_QUERIES:
        body = blocks.get(name)
        if body is None:
            findings.append(f"queries/incidents.sql: query {name} not found")
        elif "LIMIT 200" not in body:
            findings.append(f"queries/incidents.sql: {name} has no LIMIT 200 cap")


def check_named_limit_200_queries(findings):
    for rel, names in NAMED_LIMIT_200_QUERIES:
        text = (API / rel).read_text()
        blocks = dict(re.findall(r"-- name: (\w+) :\w+\n(.*?)(?=\n-- name:|\Z)", text, re.DOTALL))
        for name in names:
            body = blocks.get(name)
            if body is None:
                findings.append(f"{rel}: query {name} not found")
            elif "LIMIT 200" not in body:
                findings.append(f"{rel}: {name} has no LIMIT 200 cap")


def check_active_incidents_batched(findings):
    text = (API / "internal" / "handler" / "status_public.go").read_text()
    if "ListStatusPageIncidentUpdatesForIncidents" not in text:
        findings.append(
            "status_public.go: loadActiveIncidents no longer calls the batched "
            "ListStatusPageIncidentUpdatesForIncidents — may have regressed to one query per incident (N+1)"
        )


# Each entry: (handler file, "const = N" snippet, "call-site" snippet) for a
# flat, uniform-across-every-plan creation cap — the four added 2026-07-12
# after ListStatusPageIncidentUpdates' missing LIMIT led to auditing every
# other unbounded-creation handler in the same pass.
FLAT_CAPS = [
    ("incidents.go", "maxActiveIncidents = 100", "checkActiveIncidentCap(ctx, w, orgID)"),
    ("incidents.go", "maxUpdatesPerIncident = 100", "checkUpdateCap(ctx, w, incident.ID)"),
    ("maintenance.go", "maxMaintenanceWindows = 100", "count >= maxMaintenanceWindows"),
    ("api_keys.go", "maxAPIKeys = 100", "count >= maxAPIKeys"),
]


def check_flat_caps(findings):
    for filename, const_snippet, call_snippet in FLAT_CAPS:
        text = (API / "internal" / "handler" / filename).read_text()
        if const_snippet not in text:
            findings.append(f"{filename}: {const_snippet!r} no longer found")
        if call_snippet not in text:
            findings.append(f"{filename}: {call_snippet!r} no longer found — cap may not be wired in")


def check_pruning_wired(findings):
    worker_go = (API / "internal" / "worker" / "worker.go").read_text()
    for name in PRUNE_QUERIES:
        if f"queries.{name}(ctx)" not in worker_go:
            findings.append(f"worker.go: pruneOldPings no longer calls {name}")


def check_status_page_rate_limit(findings):
    text = (API / "internal" / "server" / "server.go").read_text()
    routes = re.findall(r'r\.With\(httprate\.LimitByIP\((\d+), time\.Minute\)\)\.Get\("(/status/[^"]*)"', text)
    if len(routes) < 3:
        findings.append(f"server.go: expected 3 IP-rate-limited /status/* routes (page, badge, monitor badge), found {len(routes)}")
    for limit, path in routes:
        if limit != "300":
            findings.append(f"server.go: {path} rate-limited at {limit}/min, doc says 300/min")


def check_blanket_org_limit(findings):
    text = (API / "internal" / "server" / "server.go").read_text()
    if "httprate.Limit(300, time.Minute, httprate.WithKeyFuncs(authOrgKey))" not in text:
        findings.append("server.go: no blanket 300/min-per-org httprate.Limit(...) found on the RequireAuth group")


def check_body_limit(findings):
    text = (API / "internal" / "server" / "server.go").read_text()
    if "http.MaxBytesReader(w, r.Body, 64*1024)" not in text:
        findings.append("server.go: no global 64 KB http.MaxBytesReader(...) found")


def check_monitor_plan_limits(findings):
    text = (API / "internal" / "billing" / "plans.go").read_text()
    totals = re.findall(r"MonitorTotal:\s*(-?\d+)", text)
    if len(totals) < 4:
        findings.append(f"plans.go: expected 4 plans with MonitorTotal set, found {len(totals)}")
    elif any(int(t) == -1 for t in totals):
        findings.append("plans.go: at least one plan has an unlimited (-1) MonitorTotal — doc says every plan is capped (10-1000)")


CHECKS = [
    check_bounded_concurrency,
    check_incident_limits,
    check_status_page_incident_limits,
    check_named_limit_200_queries,
    check_flat_caps,
    check_active_incidents_batched,
    check_pruning_wired,
    check_status_page_rate_limit,
    check_blanket_org_limit,
    check_body_limit,
    check_monitor_plan_limits,
]


def main() -> int:
    findings = []
    for check in CHECKS:
        check(findings)

    if findings:
        print("docs/reference/limits.md claims that no longer hold:")
        for f in findings:
            print(f"  - {f}")
        print(f"\n{len(findings)} drifted claim(s). Update the code (if this is a real regression) "
              "or the doc (if the change was deliberate).")
        return 1

    print("All checkable claims in docs/reference/limits.md still hold.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
