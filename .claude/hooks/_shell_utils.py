"""Shared helpers for .claude/hooks scripts."""
# Deliberately simple, best-effort shell parsing — just enough to tell a
# real subcommand invocation apart from the same text appearing inside a
# quoted string (a grep pattern, an echoed JSON payload). Not a full shell
# parser.
#
# Caught during authoring: a naive substring `grep` for "make deploy"
# anywhere in tool_input.command false-positived on `grep -n "...make
# deploy..." file` — the text was inside a quoted grep pattern, never
# actually executed. These helpers blank out quoted contents first so
# matching only sees real syntax.

import re

_DOUBLE_QUOTED = re.compile(r'"(?:[^"\\]|\\.)*"')
_SINGLE_QUOTED = re.compile(r"'[^']*'")


def strip_quotes(command: str) -> str:
    """Blank out quoted-string contents, keeping quote chars and length so a later regex's word-boundary logic still lines up."""

    def blank(m: "re.Match[str]") -> str:
        s = m.group(0)
        return s[0] + " " * (len(s) - 2) + s[-1]

    command = _DOUBLE_QUOTED.sub(blank, command)
    command = _SINGLE_QUOTED.sub(blank, command)
    return command


def segments(command: str) -> list[str]:
    """Split into logical command segments on ; && || | and newlines outside quotes, returning each segment's original unstripped text."""
    marked = strip_quotes(command)
    parts: list[str] = []
    last = 0
    i = 0
    n = len(marked)
    while i < n:
        two = marked[i : i + 2]
        if two in ("&&", "||"):
            parts.append(command[last:i])
            i += 2
            last = i
            continue
        one = marked[i]
        if one in (";", "|", "\n"):
            parts.append(command[last:i])
            i += 1
            last = i
            continue
        i += 1
    parts.append(command[last:])
    return [p.strip() for p in parts if p.strip()]
