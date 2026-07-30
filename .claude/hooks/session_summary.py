#!/usr/bin/env python3
"""SessionEnd hook — captures ADR/story/doc-worthy outcomes before the session closes."""
# Stop/SessionEnd only support "command" hooks (no "agent"/"prompt" type), so
# real judgment — read the transcript, check existing docs for duplicates,
# decide whether anything here clears the bar for a permanent doc — has to
# happen in a separate headless `claude -p` call this script spawns and
# detaches (the parent process may be exiting by the time it finishes).
# That call is scoped tightly: Bash denied, file edits allowed only under
# docs/decisions/, docs/stories/, docs/knowledge/, docs/roadmap.md, and
# docs/INDEX.md — see PERMISSION_SETTINGS below — and it never touches git,
# so any new file just shows up as untracked/modified for the next session
# to review and commit normally.
#
# A cheap local pre-filter (transcript size/line count) skips trivial
# sessions before paying for the headless call at all.
import json
import os
import subprocess  # nosec B404 - fixed argv below, never shell=True
import sys
from datetime import datetime, timezone

PROJECT_DIR = os.environ.get("CLAUDE_PROJECT_DIR", os.getcwd())
LOG_DIR = os.path.join(PROJECT_DIR, ".claude", "logs")
LOG_FILE = os.path.join(LOG_DIR, "session-summary.log")

# Below either threshold, treat the session as too small to plausibly
# contain an ADR/story-worthy decision — skip without spawning anything.
MIN_TRANSCRIPT_BYTES = 20_000
MIN_TRANSCRIPT_LINES = 40

PERMISSION_SETTINGS = {
    "permissions": {
        "allow": [
            "Read",
            "Grep",
            "Glob",
            "Edit(docs/decisions/**)",
            "Edit(docs/stories/**)",
            "Edit(docs/knowledge/**)",
            "Edit(docs/roadmap.md)",
            "Edit(docs/INDEX.md)",
        ],
        "deny": ["Bash"],
        "defaultMode": "dontAsk",
    }
}

PROMPT = """You are running as an unattended SessionEnd hook for the checkmeup \
repo, after a Claude Code session just ended. Your only input is the \
transcript below — there is no user to ask questions of, so when in doubt, \
do nothing rather than guess.

Transcript file: {transcript_path}
(It may be large — grep it for signal first: decisions, trade-offs, "ADR", \
"TODO", "should", "later", things explicitly deferred or ruled out, rather \
than reading it start to end. Read only the surrounding context of what you \
find.)

Your job: decide whether anything in this session is worth a permanent, \
durable record — and if so, record it in exactly the right existing place, \
never a new ad hoc doc.

Bar for "worth recording" (be conservative — most sessions produce nothing \
here, and that's the expected, correct outcome):
- A non-trivial technical/architectural decision that was actually made \
(not just discussed) and has a rationale someone would need later to \
understand *why*, not just *what*.
- A concrete follow-up task or feature idea that isn't already tracked \
anywhere and is substantial enough to be worth not forgetting.
- A durable project fact, gotcha, or workflow rule that a future session \
would benefit from knowing and that isn't already documented.
- Explicitly NOT: routine implementation work already self-evident from the \
diff/commit messages, anything already covered by CLAUDE.md, an existing \
ADR, an existing story, or an existing docs/knowledge/*.md snapshot.

Where to record it, matching this repo's real conventions (read one real \
example file in each location before writing — do not use \
docs/decisions/_template.md or docs/stories/_template.md verbatim, the \
actual files don't follow that template's frontmatter):
1. A decision that was actually made and acted on this session \
-> new file docs/decisions/NNN-slug.md, NNN = one past the highest number \
currently in docs/decisions/ (Glob it first). Match the exact shape of a \
recent real ADR (e.g. docs/decisions/036-flat-safety-caps.md): \
"# ADR-NNN: Title", "**Date:**", "**Status:** Accepted", "---", \
"## Context", "## Decision", "## Consequences". Then add one row to \
docs/INDEX.md's "## decisions/" table (same one-line "hook" style as the \
existing rows), and if a matching open question exists in \
docs/decisions/backlog.md, remove it there.
2. A real open question that surfaced but was NOT resolved this session \
-> append one bullet to docs/decisions/backlog.md, matching its existing \
bullet format and voice exactly (read the existing bullets first).
3. A concrete, not-yet-scoped follow-up idea -> append one bullet under the \
right Now/Next/Later section of docs/roadmap.md, matching its existing \
bullet format.
4. A durable fact/gotcha/workflow rule -> append a short note to the most \
relevant existing file under docs/knowledge/ (do not create a new \
knowledge file unless the topic genuinely has no existing home there).

Before writing anything in any of the four cases: grep docs/decisions/, \
docs/stories/, docs/decisions/backlog.md, docs/roadmap.md, docs/knowledge/, \
docs/INDEX.md, and CLAUDE.md for the same ground already being covered. If \
it is — even partially — skip; do not create a near-duplicate.

Constraints:
- Never touch git (no add/commit/push) — leave any new/changed file \
uncommitted for a human to review next session.
- At most one new file or a small number of targeted edits. If nothing \
clears the bar, make no changes at all.
- End your response with exactly one line starting with "RESULT:" \
summarizing what you did or why you did nothing, e.g. \
"RESULT: created docs/decisions/039-foo.md" or \
"RESULT: nothing significant this session".
"""


def log(line: str) -> None:
    os.makedirs(LOG_DIR, exist_ok=True)
    with open(LOG_FILE, "a", encoding="utf-8") as f:
        f.write(line + "\n")


def transcript_clears_bar(path: str) -> bool:
    try:
        size = os.path.getsize(path)
        if size < MIN_TRANSCRIPT_BYTES:
            return False
        with open(path, "r", encoding="utf-8", errors="ignore") as f:
            for i, _ in enumerate(f):
                if i + 1 >= MIN_TRANSCRIPT_LINES:
                    return True
        return False
    except OSError:
        return False


def main() -> int:
    try:
        data = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    session_id = data.get("session_id") or "unknown"
    transcript_path = data.get("transcript_path") or ""

    if not transcript_path or not os.path.isfile(transcript_path):
        return 0

    os.makedirs(LOG_DIR, exist_ok=True)
    marker = os.path.join(LOG_DIR, f".summarized-{session_id}")
    if os.path.exists(marker):
        return 0

    if not transcript_clears_bar(transcript_path):
        return 0

    open(marker, "w", encoding="utf-8").close()

    prompt = PROMPT.format(transcript_path=transcript_path)
    cmd = [
        "claude",
        "-p",
        prompt,
        "--settings",
        json.dumps(PERMISSION_SETTINGS),
        "--model",
        "sonnet",
    ]

    log_fh = open(LOG_FILE, "a", encoding="utf-8")
    log_fh.write(
        f"\n=== {datetime.now(timezone.utc).isoformat()} session={session_id} ===\n"
    )
    log_fh.flush()

    try:
        subprocess.Popen(  # nosec B603 - fixed argv, no shell; prompt is a fixed template with one path substitution, not attacker-controlled
            cmd,
            cwd=PROJECT_DIR,
            stdout=log_fh,
            stderr=subprocess.STDOUT,
            stdin=subprocess.DEVNULL,
            start_new_session=True,
        )
    except OSError as e:
        log(f"session_summary.py: failed to spawn claude -p: {e}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
