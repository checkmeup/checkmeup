#!/usr/bin/env python3
"""PreToolUse guard for Bash — blocks any real `kamal` invocation."""
# Any real `kamal` invocation targets the real production server
# (checkmeup.net) via config/deploy.yml. Only the human operator runs it,
# so this blocks unconditionally once a segment actually invokes kamal
# (segment- and quote-aware, see _shell_utils.py — so a mention of "kamal"
# inside a quoted string doesn't trigger this).
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _shell_utils import segments, strip_quotes  # noqa: E402

_KAMAL = re.compile(r"(^|\s)kamal(\s|$)")


def main() -> int:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "") or ""

    for seg in segments(command):
        if _KAMAL.search(strip_quotes(seg)):
            print(
                "Blocked: kamal targets the real production server "
                "(checkmeup.net). Only the human operator runs it — see "
                "CLAUDE.md's Don't section.",
                file=sys.stderr,
            )
            return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
