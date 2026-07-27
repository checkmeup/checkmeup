#!/usr/bin/env python3
"""PreToolUse guard for Bash — blocks docker build/buildx against config/deploy.yml."""
# That target's real production image/registry (checkmeup.net, GHCR).
# Plenty of legitimate docker build/buildx calls exist in this repo (test
# images, etc.) — only the one against the real deploy config is forbidden
# for Claude.
#
# Segment- and quote-aware (see _shell_utils.py) so this only fires on a
# real docker build/buildx invocation that references config/deploy.yml,
# not on that text appearing inside an unrelated quoted string.
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_DOCKER_BUILD = re.compile(r"(^|\s)docker\s+build(x)?\b")
_DEPLOY_YML = re.compile(r"config/deploy\.yml")


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""

    for seg in segments(command):
        stripped = strip_quotes(seg)
        if _DOCKER_BUILD.search(stripped) and _DEPLOY_YML.search(stripped):
            print(
                "Blocked: docker build/buildx against config/deploy.yml "
                "targets the real production image/registry (checkmeup.net, "
                "GHCR). Only the human operator runs this — see CLAUDE.md's "
                "Don't section.",
                file=sys.stderr,
            )
            return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
