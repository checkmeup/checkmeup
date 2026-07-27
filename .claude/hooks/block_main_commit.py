#!/usr/bin/env python3
"""PreToolUse guard for Bash — blocks `git commit` while on local main."""
# CLAUDE.md's Don't section: never commit directly on local main, always
# branch first. Checks each command segment for a real `git commit`
# invocation (not just current branch) so this never blocks unrelated
# commands run while sitting on main — see _shell_utils.py's docstring for
# why segment-based, quote-aware matching matters here.
import json
import os
import re
import subprocess  # nosec B404 - fixed argv below, never shell=True
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_GIT_COMMIT = re.compile(r"(^|\s)git\s+commit\b")


def current_branch() -> str:
    try:
        out = subprocess.run(  # nosec B603 B607 - fixed argv, no shell, no external input
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            capture_output=True,
            text=True,
            check=False,
        )
        return out.stdout.strip()
    except OSError:
        return ""


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""

    for seg in segments(command):
        if _GIT_COMMIT.search(strip_quotes(seg)) and current_branch() == "main":
            print(
                "Blocked: committing directly on main. Create/switch to a "
                'feature branch first — see CLAUDE.md\'s Don\'t section '
                '("Commit directly onto local main").',
                file=sys.stderr,
            )
            return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
