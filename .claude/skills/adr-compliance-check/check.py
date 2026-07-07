#!/usr/bin/env python3
"""Check the codebase against a handful of concrete, greppable ADR rules."""
# CLAUDE.md's Don't section rejects several architectural patterns
# outright (add Redis/a queue, use an ORM, switch payment providers, use
# Authorization for browser session auth). Each check below is scoped to
# the exact place a real violation would show up, verified against this
# repo's actual layout when authored — not a blanket repo-wide grep,
# which would drown in false positives (e.g. "Authorization" and
# "Stripe"/"LemonSqueezy" both appear constantly in blog-post prose and
# ADR-history comments that are legitimate).
import re
import sys
from pathlib import Path

BROKER_PATTERN = re.compile(r"redis|amqp|rabbitmq|nsq|kafka", re.IGNORECASE)
ORM_PATTERN = re.compile(r"gorm\.io|entgo\.io|uptrace/bun|go-xorm", re.IGNORECASE)
PAYMENT_PROVIDER_PATTERN = re.compile(r"stripe|lemonsqueezy|lemon-squeezy", re.IGNORECASE)


def check_go_mod_deps(path: str, pattern: re.Pattern, label: str):
    text = Path(path).read_text()
    hits = [line.strip() for line in text.splitlines() if pattern.search(line)]
    return [(path, label, h) for h in hits]


def check_package_json_deps(path: str, pattern: re.Pattern, label: str):
    text = Path(path).read_text()
    hits = [line.strip() for line in text.splitlines() if pattern.search(line)]
    return [(path, label, h) for h in hits]


def check_auth_init_uses_fetch(path: str):
    text = Path(path).read_text()
    m = re.search(r"async function init\(\).*?\n\}", text, re.S)
    if not m:
        return [(path, "auth.init() not found", "could not locate the function to check")]
    body = m.group(0)
    if "fetch(" not in body:
        return [(path, "auth.init() no longer uses plain fetch", body.strip()[:200])]
    if re.search(r"\bapi\.(get|post|put|patch|delete)\(", body):
        return [(path, "auth.init() calls the api.* client (401 interceptor risk)", body.strip()[:200])]
    return []


def check_no_authorization_header(src_dir: str, exclude_dir: str):
    findings = []
    for path in Path(src_dir).rglob("*"):
        if path.suffix not in (".ts", ".vue") or exclude_dir in path.parts:
            continue
        for i, line in enumerate(path.read_text().splitlines(), start=1):
            if "Authorization" in line:
                findings.append((str(path), f"line {i}", line.strip()))
    return findings


def run_checks():
    findings = {}
    findings["ADR-001 (no broker/queue)"] = check_go_mod_deps(
        "apps/api/go.mod", BROKER_PATTERN, "go.mod dependency"
    )
    findings["ADR-004 (no ORM)"] = check_go_mod_deps("apps/api/go.mod", ORM_PATTERN, "go.mod dependency")
    findings["ADR-026 (Paddle only, no Stripe/LemonSqueezy)"] = check_go_mod_deps(
        "apps/api/go.mod", PAYMENT_PROVIDER_PATTERN, "go.mod dependency"
    ) + check_package_json_deps("apps/web/package.json", PAYMENT_PROVIDER_PATTERN, "package.json dependency")
    findings["auth.init() must use plain fetch, not api.* (401-interceptor Don't rule)"] = (
        check_auth_init_uses_fetch("apps/web/src/stores/auth.ts")
    )
    findings["ADR-003 (no Authorization header for browser session auth)"] = check_no_authorization_header(
        "apps/web/src", "blog"
    )
    return findings


def report(findings) -> int:
    any_failed = False
    for rule, hits in findings.items():
        if not hits:
            print(f"OK   {rule}")
            continue
        any_failed = True
        print(f"FAIL {rule}")
        for hit in hits:
            print(f"       {hit}")
    return 1 if any_failed else 0


if __name__ == "__main__":
    sys.exit(report(run_checks()))
