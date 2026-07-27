#!/usr/bin/env python3
"""PostToolUse reminder for Bash — blocks after a real `gh pr create` invocation."""
# Makes "log hours each time a PR is created" a standing instruction
# rather than something depending on being asked. Checks each command
# segment for a real `gh pr create` invocation (segment- and quote-aware,
# see _shell_utils.py) so an unrelated `gh pr view`/`gh pr checks` call,
# or a mention of "gh pr create" inside a quoted string, never triggers
# this — previously gated only by the settings.json `if` frontmatter
# field, which fails open on commands it can't parse and fired on every
# Bash call regardless of content.
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_GH_PR_CREATE = re.compile(r"(^|\s)gh\s+pr\s+create(\s|$)")


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""

    for seg in segments(command):
        if _GH_PR_CREATE.search(strip_quotes(seg)):
            print(json.dumps({
                "decision": "block",
                "reason": "A PR was just created. Recalculate and log "
                          "today's hours now (use the log-hours skill), "
                          "then continue.",
            }))
            return 0

    return 0


if __name__ == "__main__":
    sys.exit(main())
