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
import difflib
import io
import re
import subprocess  # nosec B404 - only ever invoked with a fixed lizard/git argv below, never a shell
import sys
from pathlib import Path

GO_DIRS = ["apps/api/internal/handler", "apps/api/internal/worker"]
GO_FUNC_CCN_THRESHOLD = 15
GO_FILE_NLOC_THRESHOLD = 700

VUE_VIEW_DIR = Path("apps/web/src/views")
VUE_VIEW_LINE_THRESHOLD = 600
VUE_COMPONENT_DIR = Path("apps/web/src/components")
VUE_COMPONENT_LINE_THRESHOLD = 250

# Sibling near-duplicates: two files whose names differ only by one of these
# tokens are the same screen in two modes, so they drift together and every
# change has to be made twice. Unlike the size checks above this isn't
# outlier detection — a majority-duplicated pair is a finding at any size,
# which is exactly the class the line-count thresholds can't see (all six
# current pairs sit well under VUE_VIEW_LINE_THRESHOLD).
SIBLING_TOKENS = [("Create", "Edit")]
SIBLING_DUP_THRESHOLD = 50  # percent of normalized lines in common
# Percentage alone stops meaning anything once the real duplication is gone:
# two thin views still score ~65% on `<script setup>`, `try {`, `} finally {`
# and their ref declarations, because the denominator shrank with the
# numerator. The absolute count is what separates "same screen written twice"
# from "two files of the same shape" — see SKILL.md for the measured gap.
SIBLING_DUP_MIN_SHARED_LINES = 100

CHURN_COMMITS = 150  # how far back to look for hot spots
CHURN_TOP_N = 10
CHURN_PATHS = re.compile(r"^(apps/api/internal|apps/web/src)/.*\.(go|ts|vue)$")
CHURN_EXCLUDE = re.compile(r"(_test\.go|\.test\.ts)$")

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


def run_lizard_file_totals(paths: list[str]) -> dict[str, int]:
    # Whole-file NLOC (used for the size check) has to come from lizard's
    # plain-text per-file summary table, not --csv: --csv is function-grained
    # only, so summing it misses top-level code (struct/type/var blocks
    # outside any function) — that undercounted real files like
    # status_public.go (824 whole-file NLOC vs 549 summed from --csv) enough
    # to hide them from this check entirely. See SKILL.md.
    cmd = ["lizard", *paths, "--exclude", "*_test.go"]
    out = subprocess.run(  # nosec B603
        cmd, capture_output=True, text=True, check=True,
    ).stdout
    totals: dict[str, int] = {}
    for line in out.splitlines():
        parts = line.split()
        # Per-file summary rows are the only 6-field lines with a numeric
        # NLOC and a bare path (no "@") as the last field — per-function
        # rows are also 6 fields but their last field is a
        # "name@start-end@path" location string, so the "@" check tells
        # them apart from header/divider/totals lines too.
        if len(parts) != 6 or "@" in parts[-1] or "/" not in parts[-1]:
            continue
        nloc_str, file = parts[0], parts[-1]
        if nloc_str.isdigit():
            totals[file] = int(nloc_str)
    return totals


def go_function_complexity(rows: list[list[str]]) -> list[tuple[str, str, str, int]]:
    findings = []
    for _nloc, ccn, _, _param, _length, _loc, file, func_name, *_rest in rows:
        if file in KNOWN_EXCEPTIONS or int(ccn) <= GO_FUNC_CCN_THRESHOLD:
            continue
        start = _loc.split("@")[1].split("-")[0]
        findings.append((file, start, func_name or "(anonymous)", int(ccn)))
    return findings


def go_file_size(totals: dict[str, int]) -> list[tuple[str, int]]:
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


def _normalized(path: Path) -> list[str]:
    """Lines with whitespace stripped and blanks dropped — reformatting shouldn't move the number."""
    return [s for s in ("".join(line.split()) for line in path.read_text(errors="ignore").splitlines()) if s]


def sibling_duplication(directory: Path) -> list[tuple[str, str, int, int]]:
    """Pairs of same-screen files (Create/Edit) sharing more than the threshold of their lines."""
    findings = []
    for left_token, right_token in SIBLING_TOKENS:
        for left in sorted(directory.rglob(f"*{left_token}*.vue")):
            right = left.with_name(left.name.replace(left_token, right_token, 1))
            if not right.exists() or str(left) in KNOWN_EXCEPTIONS:
                continue
            a, b = _normalized(left), _normalized(right)
            matcher = difflib.SequenceMatcher(None, a, b)
            pct = round(matcher.ratio() * 100)
            shared = sum(block.size for block in matcher.get_matching_blocks())
            if pct > SIBLING_DUP_THRESHOLD and shared >= SIBLING_DUP_MIN_SHARED_LINES:
                findings.append((str(left), str(right), pct, shared))
    return findings


def churn_hot_spots() -> list[tuple[str, int]]:
    """Most-touched source files in recent history — where deepening pays off first."""
    # Fixed argv, no shell, no external input.
    out = subprocess.run(  # nosec B603 B607
        ["git", "log", "--format=", "--name-only", f"-{CHURN_COMMITS}"],
        capture_output=True, text=True, check=False,
    ).stdout
    counts: dict[str, int] = {}
    for line in out.splitlines():
        if CHURN_PATHS.match(line) and not CHURN_EXCLUDE.search(line):
            counts[line] = counts.get(line, 0) + 1
    return sorted(counts.items(), key=lambda kv: -kv[1])[:CHURN_TOP_N]


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
    go_file_totals = run_lizard_file_totals(GO_DIRS)
    exit_code = 0

    if print_section(
        f"Go function complexity (CCN > {GO_FUNC_CCN_THRESHOLD})",
        sorted(go_function_complexity(go_rows), key=lambda f: -f[3]),
        lambda r: f"{r[0]}:{r[1]}: {r[2]} — CCN {r[3]}",
    ):
        exit_code = 1

    if print_section(
        f"Go handler/worker files (non-comment lines > {GO_FILE_NLOC_THRESHOLD})",
        sorted(go_file_size(go_file_totals), key=lambda f: -f[1]),
        lambda r: f"{r[0]}: {r[1]} non-comment lines",
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

    if print_section(
        f"Sibling near-duplicates ({VUE_VIEW_DIR}, >{SIBLING_DUP_THRESHOLD}% "
        f"and >={SIBLING_DUP_MIN_SHARED_LINES} shared lines)",
        sorted(sibling_duplication(VUE_VIEW_DIR), key=lambda f: -f[3]),
        lambda r: f"{r[0]} / {Path(r[1]).name}: {r[3]} shared lines ({r[2]}%)",
    ):
        exit_code = 1

    # Context, never a failure: churn says which findings to fix first, and a
    # hot file with no finding is not a defect.
    print_section(
        f"Churn hot spots (last {CHURN_COMMITS} commits — prioritize findings that overlap)",
        churn_hot_spots(),
        lambda r: f"{r[0]}: {r[1]} commits",
    )

    if exit_code:
        print("Findings above are candidates for a closer look, not automatic failures — see SKILL.md.")
    else:
        print("Nothing over threshold.")
    return exit_code


if __name__ == "__main__":
    sys.exit(report())
