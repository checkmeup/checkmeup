#!/usr/bin/env python3
"""PreToolUse guard for Bash — blocks `gh pr create` until hours are logged on this branch."""
# Hours must land in the *same* PR as the work, not a follow-up
# docs/hours-* PR — that was the pattern before 2026-07-27, and it meant
# every PR needed a second PR just to log the first one's time. Checks
# whether a docs/reports/hours/YYYY-MM.md log was touched by any commit on the current branch
# since it diverged from origin/main; if so, the branch already carries
# an hours entry and gh pr create is allowed through. If not, blocks and
# tells Claude to log hours (log-hours skill) and commit it here first —
# segment- and quote-aware (see _shell_utils.py) so this only fires on a
# real `gh pr create` invocation, never on that text inside a quoted
# string.
import json
import os
import re
import subprocess  # nosec B404 - fixed argv below, never shell=True
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_GH_PR_CREATE = re.compile(r"(^|\s)gh\s+pr\s+create(\s|$)")
# The hours log is one file per month under docs/reports/hours/ (README.md
# there is the index, not a log — touching it alone doesn't count as
# logging hours).
_HOURS_LOG = re.compile(r"^docs/reports/hours/\d{4}-\d{2}\.md$")


def hours_logged_on_this_branch() -> bool:
    merge_base = subprocess.run(  # nosec B603 B607 - fixed argv, no shell, no external input
        ["git", "merge-base", "origin/main", "HEAD"],
        capture_output=True,
        text=True,
        check=False,
    )
    if merge_base.returncode != 0 or not merge_base.stdout.strip():
        return True  # can't determine branch point — fail open, don't block on a git error

    diff = subprocess.run(  # nosec B603 B607 - fixed argv, no shell, no external input
        ["git", "diff", "--name-only", merge_base.stdout.strip(), "HEAD"],
        capture_output=True,
        text=True,
        check=False,
    )
    return any(_HOURS_LOG.match(path) for path in diff.stdout.splitlines())


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""

    for seg in segments(command):
        if _GH_PR_CREATE.search(strip_quotes(seg)) and not hours_logged_on_this_branch():
            print(
                "Blocked: log today's hours on this branch before creating the PR "
                "(use the log-hours skill, commit docs/reports/hours/YYYY-MM.md + docs/reports/*.md "
                "here), so the hours entry ships in the same PR as the work — "
                "not a separate follow-up PR. Then retry gh pr create.",
                file=sys.stderr,
            )
            return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
