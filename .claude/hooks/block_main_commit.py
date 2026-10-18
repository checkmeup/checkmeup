#!/usr/bin/env python3
"""PreToolUse guard for Bash — blocks `git commit` while on local main."""
# CLAUDE.md's Don't section: never commit directly on local main, always
# branch first. Checks each command segment for a real `git commit`
# invocation (not just current branch) so this never blocks unrelated
# commands run while sitting on main — see _shell_utils.py's docstring for
# why segment-based, quote-aware matching matters here.
#
# The branch is read in the repo the commit runs in: the session's cwd,
# moved by any earlier `cd` in the command, or a `git -C <path>`. A
# `git switch` or `git checkout -b` earlier in the same command counts, since
# the hook runs before any of it does.
import json
import os
import re
import subprocess  # nosec B404 - fixed argv below, never shell=True
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_GIT = re.compile(r"(?:^|\s)git((?:\s+-[Cc]\s+\S+|\s+--\S+)*)\s+(\S+)(.*)")
_DASH_C = re.compile(r"-C\s+(\S+)")
_CD = re.compile(r"^cd\s+(\S+)\s*$")
_SWITCH_CREATE = re.compile(r"^\s+(?:-c|-C|--create|--force-create)\s+(\S+)")
_CHECKOUT_CREATE = re.compile(r"^\s+(?:-b|-B)\s+(\S+)")
_PLAIN_ARG = re.compile(r"^\s+([^-\s]\S*)\s*$")


def current_branch(repo: str) -> str:
    try:
        out = subprocess.run(  # nosec B603 B607 - fixed argv, no shell, no external input
            ["git", "-C", repo, "rev-parse", "--abbrev-ref", "HEAD"],
            capture_output=True,
            text=True,
            check=False,
        )
        return out.stdout.strip()
    except OSError:
        return ""


def _resolve(cwd: str, path: str) -> str:
    return os.path.normpath(os.path.join(cwd, os.path.expanduser(path.strip("'\""))))


def _switched_to(sub: str, rest: str) -> str:
    """The branch a `git switch`/`git checkout` segment moves to, or "" if none is clear."""
    if sub == "switch":
        m = _SWITCH_CREATE.match(rest) or _PLAIN_ARG.match(rest)
    elif sub == "checkout":
        m = _CHECKOUT_CREATE.match(rest) or _PLAIN_ARG.match(rest)
    else:
        return ""
    return m.group(1).strip("'\"") if m else ""


def commit_on_main(command: str, cwd: str) -> bool:
    switched: dict[str, str] = {}
    for seg in segments(command):
        stripped = strip_quotes(seg)
        cd = _CD.match(seg)
        if cd:
            cwd = _resolve(cwd, cd.group(1))
            continue
        git = _GIT.search(stripped)
        if not git:
            continue
        # Same length as seg, so the spans read the original, quotes included.
        paths = _DASH_C.findall(seg[git.start(1) : git.end(1)])
        repo = _resolve(cwd, paths[-1]) if paths else cwd
        sub, rest = git.group(2), seg[git.start(3) : git.end(3)]
        if sub == "commit":
            branch = switched.get(repo) or current_branch(repo)
            if branch == "main":
                return True
        elif branch := _switched_to(sub, rest):
            switched[repo] = branch
    return False


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""
    cwd = data.get("cwd") or os.getcwd()

    if commit_on_main(command, cwd):
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
