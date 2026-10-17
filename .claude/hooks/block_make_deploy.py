#!/usr/bin/env python3
"""PreToolUse guard for Bash — blocks the deploy and ghcr-clean tasks (mise or make), dry-run flags included."""
# CLAUDE.md's Don't section: only the human operator runs `mise run deploy` /
# `mise run ghcr-clean`, which target the real production server/registry
# (checkmeup.net, GHCR). The tasks lived in a Makefile until 2026-10-17, so
# `make deploy` / `make ghcr-clean` stay blocked too, for an old checkout.
#
# A dry-run flag does not make this safe: `make -n deploy` has actually
# triggered a real deploy before (docs/incidents/2026-07-02-…), so flags
# around the task name still need to match below. Any mise subcommand that
# names one of the tasks is blocked (run, r, watch, the bare `mise deploy`
# shorthand), except `mise tasks info|ls|deps|edit`, which only read it.
#
# Segment- and quote-aware (see _shell_utils.py) so this only fires on a
# real invocation, never on that text appearing inside an unrelated quoted
# string (a grep pattern, a comment).
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_MAKE_DEPLOY = re.compile(r"(^|\s)make\s+(-\S+\s+)*(deploy|ghcr-clean)\b")
_MISE_DEPLOY = re.compile(r"(^|\s)mise\s+(.*\s)?(deploy|ghcr-clean)(\s|$)")
_MISE_READ_ONLY = re.compile(r"(^|\s)mise\s+(-\S+\s+)*tasks\s+(info|ls|deps|edit)\b")


def _is_deploy(seg: str) -> bool:
    if _MAKE_DEPLOY.search(seg):
        return True
    return bool(_MISE_DEPLOY.search(seg)) and not _MISE_READ_ONLY.search(seg)


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""

    for seg in segments(command):
        if _is_deploy(strip_quotes(seg)):
            print(
                "Blocked: the deploy/ghcr-clean tasks target the real production "
                "server/registry (checkmeup.net, GHCR). Only the human "
                "operator runs these — see CLAUDE.md's Don't section.",
                file=sys.stderr,
            )
            return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
