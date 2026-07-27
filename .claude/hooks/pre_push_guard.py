#!/usr/bin/env python3
"""PreToolUse guard for Bash — checks each real `git push` for force-push and secret-leak risks."""
# Segment- and quote-aware (see _shell_utils.py). Guards three things:
#
#   1. Bare --force is blocked in favor of --force-with-lease (pr-merge's
#      own convention — never clobber a push you haven't fetched).
#   2. Force-pushing main is always blocked — main is only ever
#      fast-forwarded (pr-merge's rebase-only convention). Requires "main"
#      to actually be the pushed ref (origin main / a :main refspec /
#      trailing bare main), not just anywhere in the command — a branch
#      literally named "release/main-2" should not trip this.
#   3. secrets-scan runs against the full tracked tree before any push,
#      turning the "good pre-flight before pr-merge" advice into a real
#      gate instead of a step that depends on remembering to run it.
import json
import os
import re
import subprocess  # nosec B404 - fixed argv below, never shell=True
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_GIT_PUSH = re.compile(r"(^|\s)git\s+push\b")
_BARE_FORCE = re.compile(r"(^|\s)--force(\s|$)")
_FORCE_ANY = re.compile(r"--force(-with-lease)?\b")
_TARGETS_MAIN = re.compile(r"(\borigin\s+main\b|:main\b|\smain$)")


def run_secrets_scan() -> str | None:
    project_dir = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
    scan_py = os.path.join(project_dir, ".claude", "skills", "secrets-scan", "scan.py")
    # scan_py is CLAUDE_PROJECT_DIR (harness-set, not attacker input) joined
    # with fixed literal path segments, never passed through a shell.
    result = subprocess.run(  # nosec B603 B607 - fixed argv, no shell, no external input
        [sys.executable, scan_py, "tree"],  # nosemgrep: dangerous-subprocess-use-tainted-env-args
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        return result.stdout + result.stderr
    return None


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""

    push_segments = [seg for seg in segments(command) if _GIT_PUSH.search(strip_quotes(seg))]
    if not push_segments:
        return 0

    for seg in push_segments:
        stripped = strip_quotes(seg)

        if _BARE_FORCE.search(stripped) and "--force-with-lease" not in stripped:
            print(
                "Blocked: use --force-with-lease instead of bare --force, so "
                "a push you haven't fetched can't be silently overwritten — "
                "see the pr-merge skill.",
                file=sys.stderr,
            )
            return 2

        if _FORCE_ANY.search(stripped) and _TARGETS_MAIN.search(stripped):
            print(
                "Blocked: never force-push main — it should only ever be "
                "fast-forwarded (git merge --ff-only), per the pr-merge "
                "skill's rebase-only convention.",
                file=sys.stderr,
            )
            return 2

    scan_output = run_secrets_scan()
    if scan_output is not None:
        print(
            "Blocked: secrets-scan found a hit before this push. Run "
            "`python3 .claude/skills/secrets-scan/scan.py tree` to see "
            "details and fix before pushing.",
            file=sys.stderr,
        )
        return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
