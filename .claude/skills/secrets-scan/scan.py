#!/usr/bin/env python3
"""Scan for credential-shaped strings and accidentally-tracked secrets."""
# Covers staged changes (default) or the full git-tracked tree (--tree).
#
# Deliberately scoped to git-tracked/staged content only — it never reads
# the working tree's actual .env files, so legitimate local secrets in
# apps/api/.env or apps/web/.env are never touched or at risk of being
# echoed anywhere. Patterns match specific known credential *shapes*
# (provider-prefixed IDs, private-key headers, embedded DB creds), not a
# generic "any assigned string near the word secret" heuristic — that
# generic form flags apps/api/.env.example's placeholder values (empty, or
# "change-me-in-production...", "your-codacy-api-token") constantly, which
# trains people to ignore the tool. If a new provider is integrated, add
# its key format below rather than loosening this to a generic heuristic.
import re
import subprocess  # nosec B404 - only ever invoked with a fixed git argv below, never a shell
import sys

# Patterns marked context_exempt=True are skipped in test files, docs, and
# .example files — this repo's checked-in dev-only Postgres credential
# (postgres://checkmeup:checkmeup@..., matching docker-compose.yml) shows
# up in .env.example, CI config, and *_test.go by design (same precedent
# as CLAUDE.md's "ignore cookies in *_test.go" — synthetic fixtures, not
# leaks), and docs/reference/deploy.md's DB URL is a documented
# <placeholder>, not a real value. Provider-token shapes (Twilio/GitHub/
# AWS/private-key/Telegram) have no such legitimate fixture use anywhere
# in this repo (verified empty when authoring this scan) and stay strict
# everywhere, docs included.
PATTERNS = [
    ("Twilio Account SID", re.compile(r"\bAC[0-9a-fA-F]{32}\b"), False),
    ("Twilio API Key SID", re.compile(r"\bSK[0-9a-fA-F]{32}\b"), False),
    ("GitHub token", re.compile(r"\b(?:ghp|gho|ghu|ghs|ghr|github_pat)_[A-Za-z0-9_]{20,}\b"), False),
    ("AWS access key ID", re.compile(r"\bAKIA[0-9A-Z]{16}\b"), False),
    ("Resend API key", re.compile(r"\bre_[A-Za-z0-9]{20,}\b"), False),
    ("Private key block", re.compile(r"-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----"), False),
    ("Slack webhook URL", re.compile(r"https://hooks\.slack\.com/services/[A-Za-z0-9/]+"), True),
    ("DB URL with embedded credentials", re.compile(r"(?:postgres|postgresql|mysql)://[^:/\s\"']+:[^@/\s\"']+@"), True),
    # Telegram bot token: numeric bot ID, colon, 35-char alnum/underscore/dash secret.
    ("Telegram bot token", re.compile(r"\b\d{6,10}:[A-Za-z0-9_-]{35}\b"), False),
]

CONTEXT_EXEMPT_FILE = re.compile(r"_test\.go$|\.test\.tsx?$|\.md$|\.example$")

# Files that should never be tracked, regardless of content.
FORBIDDEN_FILENAME = re.compile(r"(^|/)\.env(\..+)?$|\.pem$|\.key$|id_rsa$|id_ed25519$")
# .env.example / .env.*.example are templates, always OK.
FILENAME_EXEMPT = re.compile(r"\.env(\..+)?\.example$")
BINARY_EXT = re.compile(
    r"\.(png|jpe?g|gif|ico|woff2?|ttf|otf|pdf|zip|gz|lock|webp|bmp)$", re.IGNORECASE
)


def run(cmd):
    # cmd is always a literal argv list built by this file (git subcommands +
    # filenames from `git ls-files`/`git diff`, never shell=True or
    # externally-supplied input), so there's no shell-injection surface.
    return subprocess.run(  # nosec B603
        cmd, capture_output=True, text=True, check=False,
        encoding="utf-8", errors="replace",
    )


def check_filenames(paths):
    findings = []
    for p in paths:
        if FILENAME_EXEMPT.search(p):
            continue
        if FORBIDDEN_FILENAME.search(p):
            findings.append(p)
    return findings


# The one documented, non-secret local-dev Postgres credential (matches
# docker-compose.yml) — appears by design in .env.example, CI config, and
# test fixtures. Exempt everywhere, same reasoning as CLAUDE.md's "ignore
# cookies in *_test.go": a known synthetic value, not a leak.
KNOWN_DEV_CREDENTIAL = "checkmeup:checkmeup"


def scan_lines(lines_with_loc):
    """lines_with_loc: iterable of (file, lineno, text)."""
    findings = []
    for file, lineno, text in lines_with_loc:
        exempt_file = bool(CONTEXT_EXEMPT_FILE.search(file))
        for name, pattern, context_exempt in PATTERNS:
            m = pattern.search(text)
            if not m:
                continue
            if KNOWN_DEV_CREDENTIAL in m.group(0):
                continue
            if context_exempt and exempt_file:
                continue
            findings.append((file, lineno, name, m.group(0)))
    return findings


def scan_staged():
    files = run(["git", "diff", "--cached", "--name-only", "--diff-filter=ACM"]).stdout.splitlines()
    filename_findings = check_filenames(files)

    lines = []
    diff = run(["git", "diff", "--cached", "-U0"]).stdout.splitlines()
    current_file = None
    current_line = None
    for line in diff:
        if line.startswith("+++ b/"):
            current_file = line[6:]
            continue
        if line.startswith("@@"):
            m = re.search(r"\+(\d+)", line)
            current_line = int(m.group(1)) if m else None
            continue
        if line.startswith("+") and not line.startswith("+++"):
            if current_file and current_line is not None:
                lines.append((current_file, current_line, line[1:]))
                current_line += 1
    content_findings = scan_lines(lines)
    return filename_findings, content_findings


def scan_tree():
    files = run(["git", "ls-files"]).stdout.splitlines()
    filename_findings = check_filenames(files)

    lines = []
    for f in files:
        if BINARY_EXT.search(f):
            continue
        show = run(["git", "show", f"HEAD:{f}"])
        # Skip binary/unreadable files quietly.
        if show.returncode != 0 or "�" in show.stdout:
            continue
        for i, text in enumerate(show.stdout.splitlines(), start=1):
            lines.append((f, i, text))
    content_findings = scan_lines(lines)
    return filename_findings, content_findings


def report(filename_findings, content_findings):
    ok = True
    if filename_findings:
        ok = False
        print("Forbidden files staged/tracked (should never be committed):")
        for f in filename_findings:
            print(f"  {f}")
        print()
    if content_findings:
        ok = False
        print("Credential-shaped strings found:")
        for file, lineno, name, match in content_findings:
            redacted = match[:6] + "…" + match[-4:] if len(match) > 12 else "…"
            print(f"  {file}:{lineno}  [{name}]  {redacted}")
        print()
    if ok:
        print("Clean — no forbidden files, no credential-shaped strings.")
        return 0
    return 1


if __name__ == "__main__":
    mode = sys.argv[1] if len(sys.argv) > 1 else "staged"
    if mode == "tree":
        sys.exit(report(*scan_tree()))
    elif mode == "staged":
        sys.exit(report(*scan_staged()))
    else:
        print("usage: scan.py [staged|tree]", file=sys.stderr)
        sys.exit(2)
