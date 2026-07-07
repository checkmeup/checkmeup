#!/usr/bin/env python3
"""Flag server.go routes with no rate-limit coverage."""
# Checks route registrations in apps/api/internal/server/server.go for
# coverage — neither their own httprate wrapper nor an enclosing
# r.Use(httprate...) on a parent r.Route/r.Group block.
#
# Chi propagates middleware registered via r.Use() down through nested
# r.Route()/r.Group() blocks (Route mounts a sub-router, but the parent's
# middleware still wraps the eventually-matched handler) — so this walks
# brace depth, tracking a "rate-limited" flag per lexical scope that child
# scopes inherit, same as the actual routing behavior.
#
# Known accepted exceptions (see SKILL.md) are reported separately, not as
# failures.
import re
import sys
from pathlib import Path

# A route call is either "r.Verb(" directly (the router var "r", not e.g.
# "r.Header.Get(" — the negative lookbehind excludes that), or ").Verb("
# at the end of a chained r.With(...).Verb(...).
VERB_RE = re.compile(r'(?<![A-Za-z0-9_])r\.(?:Get|Post|Put|Patch|Delete)\(|\)\.(?:Get|Post|Put|Patch|Delete)\(')
PATH_RE = re.compile(r'\.(?:Get|Post|Put|Patch|Delete)\("([^"]*)"')
BLOCK_OPENERS = ("r.Route(", "r.Group(")
KNOWN_EXCEPTIONS = {
    "/health": "load-balancer/monitoring health check — not user-reachable data",
    "/*": "static file serving (SPA catch-all) — no per-request cost, not a data endpoint",
}


def is_block_open(stripped: str) -> bool:
    return any(op in stripped for op in BLOCK_OPENERS) and stripped.endswith("{")


def handle_use(lines, i, n, stack) -> int:
    """r.Use(...) — if it mentions httprate, mark the current scope covered."""
    buf = lines[i].strip()
    j = i
    while buf.count("(") > buf.count(")") and j + 1 < n:
        j += 1
        buf += lines[j].strip()
    if "httprate" in buf:
        stack[-1] = True
    return j + 1


def handle_route(lines, i, n, stack, findings, exceptions) -> int:
    """A route registration — may span multiple lines (r.With(...).Verb(...))."""
    buf = lines[i].strip()
    j = i
    while not VERB_RE.search(buf) and j + 1 < n:
        j += 1
        buf += " " + lines[j].strip()
    path_match = PATH_RE.search(buf)
    route_path = path_match.group(1) if path_match else "?"
    covered = "httprate" in buf or stack[-1]
    if not covered:
        entry = (i + 1, route_path, buf.strip())
        bucket = exceptions if route_path in KNOWN_EXCEPTIONS else findings
        bucket.append(entry)
    return j + 1


def step(lines, i, n, stack, findings, exceptions) -> int:
    stripped = lines[i].strip()

    if is_block_open(stripped):
        stack.append(stack[-1])
        return i + 1
    if stripped == "})":
        if len(stack) > 1:
            stack.pop()
        return i + 1
    if stripped.startswith("r.Use("):
        return handle_use(lines, i, n, stack)
    if stripped.startswith("r.With(") or VERB_RE.search(stripped):
        return handle_route(lines, i, n, stack, findings, exceptions)
    return i + 1


def scan(lines):
    stack = [False]  # covered-flag per lexical scope, root = uncovered
    findings = []
    exceptions = []
    i, n = 0, len(lines)
    while i < n:
        i = step(lines, i, n, stack, findings, exceptions)
    return findings, exceptions


def report(findings, exceptions) -> int:
    if exceptions:
        print("Known exceptions (not flagged as failures):")
        for lineno, route_path, _ in exceptions:
            print(f"  server.go:{lineno}  {route_path}  — {KNOWN_EXCEPTIONS[route_path]}")
        print()

    if findings:
        print("Routes with NO rate-limit coverage:")
        for lineno, route_path, src in findings:
            print(f"  server.go:{lineno}  {route_path}")
            print(f"    {src}")
        print(f"\n{len(findings)} uncovered route(s).")
        return 1

    print("All routes covered (own httprate wrapper or an enclosing r.Use(httprate...)).")
    return 0


def main(path: str) -> int:
    lines = Path(path).read_text().splitlines()
    findings, exceptions = scan(lines)
    return report(findings, exceptions)


if __name__ == "__main__":
    target = sys.argv[1] if len(sys.argv) > 1 else "apps/api/internal/server/server.go"
    sys.exit(main(target))
