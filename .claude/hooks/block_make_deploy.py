#!/usr/bin/env python3
"""PreToolUse guard for Bash — blocks `make deploy`/`make ghcr-clean`, dry-run flags included."""
# CLAUDE.md's Don't section: only the human operator runs `make deploy` /
# `make ghcr-clean`, which target the real production server/registry
# (checkmeup.net, GHCR).
#
# A dry-run flag does not make this safe: `make -n deploy` has actually
# triggered a real deploy before, because the deploy recipe references
# $(MAKE) for the post-deploy ghcr-clean step, and GNU Make always executes
# a recipe line containing $(MAKE)/${MAKE}, even under -n — so flags
# appearing between `make` and the target still need to match below.
#
# Segment- and quote-aware (see _shell_utils.py) so this only fires on a
# real `make deploy`/`make ghcr-clean` invocation, never on that text
# appearing inside an unrelated quoted string (a grep pattern, a comment).
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_MAKE_DEPLOY = re.compile(r"(^|\s)make\s+(-\S+\s+)*(deploy|ghcr-clean)\b")


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""

    for seg in segments(command):
        if _MAKE_DEPLOY.search(strip_quotes(seg)):
            print(
                "Blocked: make deploy/ghcr-clean targets the real production "
                "server/registry (checkmeup.net, GHCR). Only the human "
                "operator runs these — see CLAUDE.md's Don't section.",
                file=sys.stderr,
            )
            return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
