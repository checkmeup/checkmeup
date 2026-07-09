#!/usr/bin/env python3
"""Flag Go/Vue functions and files exceeding this repo's size/complexity thresholds."""
# Objective proxies for SRP violations (a function or component quietly
# taking on more than one job). Two checks, both against real gaps in this
# repo's current size distribution when this audit was authored, not
# arbitrary round numbers:
#
#  1. Go function cyclomatic complexity > 15 (lizard's own conventional
#     default) in apps/api/internal/handler and .../worker — the two
#     packages carrying most of the business logic.
#  2. File size as a "god object" proxy: Go handler/worker files, and
#     Vue views/components, sized well outside the rest of their peers.
#     A big file isn't proof of a violation, but every one found here is
#     worth a look — see SKILL.md for how to read a finding.
import csv
import io
import subprocess  # nosec B404 - only ever invoked with a fixed lizard argv below, never a shell
import sys
from collections import defaultdict
from pathlib import Path

GO_DIRS = ["apps/api/internal/handler", "apps/api/internal/worker"]
GO_FUNC_CCN_THRESHOLD = 15
GO_FILE_NLOC_THRESHOLD = 700

VUE_VIEW_DIR = Path("apps/web/src/views")
VUE_VIEW_LINE_THRESHOLD = 600
VUE_COMPONENT_DIR = Path("apps/web/src/components")
VUE_COMPONENT_LINE_THRESHOLD = 250

# file -> rationale, verified by reading the file, not accepted on size
# alone. Empty until a finding is deliberately triaged as "large but
# cohesive" — see SKILL.md before adding to this.
KNOWN_EXCEPTIONS = {}


def run_lizard_csv(paths: list[str]) -> list[list[str]]:
    # cmd is always this literal argv (lizard + fixed dirs/flags below, never
    # shell=True or externally-supplied input), so there's no injection surface.
    cmd = ["lizard", *paths, "--exclude", "*_test.go", "--csv"]
    out = subprocess.run(  # nosec B603
        cmd, capture_output=True, text=True, check=True,
    ).stdout
    return [row for row in csv.reader(io.StringIO(out)) if row]


def go_function_complexity(rows: list[list[str]]) -> list[tuple[str, str, str, int]]:
    findings = []
    for _nloc, ccn, _, _param, _length, _loc, file, func_name, *_rest in rows:
        if file in KNOWN_EXCEPTIONS or int(ccn) <= GO_FUNC_CCN_THRESHOLD:
            continue
        start = _loc.split("@")[1].split("-")[0]
        findings.append((file, start, func_name or "(anonymous)", int(ccn)))
    return findings


def go_file_size(rows: list[list[str]]) -> list[tuple[str, int]]:
    totals: dict[str, int] = defaultdict(int)
    for nloc, _ccn, _, _param, _length, _loc, file, *_rest in rows:
        totals[file] += int(nloc)
    return [(f, n) for f, n in totals.items() if n > GO_FILE_NLOC_THRESHOLD and f not in KNOWN_EXCEPTIONS]


def vue_file_size(directory: Path, threshold: int, exclude_dirs: tuple = ()) -> list[tuple[str, int]]:
    findings = []
    for path in sorted(directory.rglob("*.vue")):
        if any(part in exclude_dirs for part in path.parts):
            continue
        n = sum(1 for _ in path.open())
        if n > threshold and str(path) not in KNOWN_EXCEPTIONS:
            findings.append((str(path), n))
    return findings


def print_section(title: str, findings: list, fmt) -> bool:
    # CodeQL py/clear-text-logging-sensitive-data false positive: title is
    # always one of the 4 hardcoded audit-section descriptions built in
    # report() below (e.g. "Go function complexity (CCN > 15)") — never
    # derived from user input, a secret, or any external source.
    print(f"## {title}")  # lgtm[py/clear-text-logging-sensitive-data]
    if not findings:
        print("  none")
        print()
        return False
    for row in findings:
        print(f"  {fmt(row)}")
    print()
    return True


def report() -> int:
    go_rows = run_lizard_csv(GO_DIRS)
    exit_code = 0

    if print_section(
        f"Go function complexity (CCN > {GO_FUNC_CCN_THRESHOLD})",
        sorted(go_function_complexity(go_rows), key=lambda f: -f[3]),
        lambda r: f"{r[0]}:{r[1]}: {r[2]} — CCN {r[3]}",
    ):
        exit_code = 1

    if print_section(
        f"Go handler/worker files (logical lines > {GO_FILE_NLOC_THRESHOLD})",
        sorted(go_file_size(go_rows), key=lambda f: -f[1]),
        lambda r: f"{r[0]}: {r[1]} logical lines",
    ):
        exit_code = 1

    if print_section(
        f"Vue views ({VUE_VIEW_DIR}, lines > {VUE_VIEW_LINE_THRESHOLD})",
        sorted(vue_file_size(VUE_VIEW_DIR, VUE_VIEW_LINE_THRESHOLD), key=lambda f: -f[1]),
        lambda r: f"{r[0]}: {r[1]} lines",
    ):
        exit_code = 1

    if print_section(
        f"Vue components ({VUE_COMPONENT_DIR}, excl. ui/, lines > {VUE_COMPONENT_LINE_THRESHOLD})",
        sorted(vue_file_size(VUE_COMPONENT_DIR, VUE_COMPONENT_LINE_THRESHOLD, exclude_dirs=("ui",)), key=lambda f: -f[1]),
        lambda r: f"{r[0]}: {r[1]} lines",
    ):
        exit_code = 1

    if exit_code:
        print("Findings above are candidates for a closer look, not automatic failures — see SKILL.md.")
    else:
        print("Nothing over threshold.")
    return exit_code


if __name__ == "__main__":
    sys.exit(report())
