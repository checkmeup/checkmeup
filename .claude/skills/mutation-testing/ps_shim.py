#!/usr/bin/env python3
"""Minimal `ps` replacement for containers that ship without procps.

Stryker's process cleanup (via tree-kill) only ever calls
`ps -o pid --no-headers --ppid <pid>` to find a test runner's children, so
answering exactly that from /proc is enough. Installed onto PATH as `ps` by
run.sh when the real one is missing.
"""
import os
import sys


def child_pids(ppid: str) -> list[str]:
    """PIDs whose parent is ppid, read straight from /proc."""
    found = []
    for entry in os.listdir("/proc"):
        if not entry.isdigit():
            continue
        try:
            with open(f"/proc/{entry}/stat", encoding="utf-8") as handle:
                # comm can contain spaces and parens, so split after the last
                # ')': the remaining fields are state, ppid, pgrp, ...
                fields = handle.read().rsplit(")", 1)[1].split()
            if fields[1] == ppid:
                found.append(entry)
        except (OSError, IndexError):
            continue
    return found


def main() -> int:
    argv = sys.argv[1:]
    if "--ppid" not in argv:
        return 0
    pids = child_pids(argv[argv.index("--ppid") + 1])
    # Real ps exits non-zero when nothing matched, and tree-kill relies on it:
    # its `code != 0` branch is the "no more children" path. Exiting 0 with
    # empty stdout instead sends it into "".match(/\d+/g).forEach and crashes
    # the whole Stryker run.
    if not pids:
        return 1
    print("\n".join(pids))
    return 0


if __name__ == "__main__":
    sys.exit(main())
