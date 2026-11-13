# This repo is set up as any project is: its Git hooks in `.githooks/` run `hathi`

Date: 2026-11-13

The repo is set up for bunderlog's plugins the way the setup skill writes it, and nothing in that
setup lives outside the commit. The Git hooks are the three committed scripts in `.githooks/`
(`commit-msg`, `pre-commit`, `pre-push`), each running `hathi check <hook>` with `.hathi.yml`'s
settings; `.hathi.yml` turns every Check on with its defaults except the Conventional Commits
types and its 140-character line limit. Secrets in a commit are checked there, by `hathi`'s own
check at commit time, not by a hook of this repo's own and not at push time. Which hooks run what,
and how a clone turns them on, is in [dev-environment](dev-environment.md).

Claude Code's committed settings are `.claude/settings.json`: the project's `PreToolUse` hooks in
`.claude/hooks/`, the deny list for secret files, the plugins the project enables, and
`enableWorkflows` for louie's Workflows. `/.claude/worktrees/`, where louie builds an Issue, is
ignored.

Nothing an earlier baloo wrote is left: the setup run of 2026-11-13 removed `.claude/baloo.yml`,
the untracked config of baloo's own session-start Git hooks, which baloo no longer reads.

## Considered options

- The baloo plugin's Git hooks with `.claude/baloo.yml` (until 2026-11-09): see
  [dev-environment](dev-environment.md), which records why the committed `.githooks/` scripts
  replaced them.
