---
name: pr-merge
description: Merge an open GitHub PR into main once CI passes, following this repo's rebase-only fast-forward convention (no gh pr merge, no squash, no merge commits). Use when asked to "merge PR #N", "merge this PR", or "merge it once CI passes".
---

# PR merge

Merges to `main` in this repo go through PR, but the actual merge is done
with plain git — `gh pr merge` does **not** work here (the GitHub PAT in
use lacks merge permission on the repo), and merge commits/squash are
disallowed so `main`'s log stays a straight line. See CLAUDE.md's
"Merging to `main`" convention.

Never commit or merge directly onto a local `main` that hasn't been
pushed — always end by pushing the fast-forwarded `main` to `origin`.

Three `PreToolUse` hooks now enforce parts of this deterministically:
`block_main_commit.py` blocks any `git commit` while on local `main`,
and `pre_push_guard.py` blocks a bare `--force` (use `--force-with-lease`,
per step 3 below) and any force-push targeting `main`, and runs
`secrets-scan` before every push. None of that changes this skill's
steps — they already follow those rules — but a push failing here with a
hook error rather than a git error is expected, not a sign something else
broke.

## Steps

Takes a PR number or URL as input. Below, `<N>` is the PR number and
`<branch>` is its head branch (get both from `gh pr view`).

**1. Resolve the PR and check CI.**

```bash
gh pr view <N> --json number,headRefName,state,url
gh pr checks <N>
```

- If `state` is already `MERGED`, stop and say so.
- If any check is `fail`, stop and report which one — do not merge.
- If checks are still `pending`, report that back and ask whether to wait
  rather than polling in a loop.

**2. Sync and check whether the branch needs rebasing onto `main`.**

```bash
git fetch origin
git log --oneline -1 origin/main
git log --oneline -1 <branch>
```

If `<branch>` already contains `origin/main`'s tip (i.e. it was branched
from current main and has no new commits from main to absorb), skip
straight to step 4 — a rebase would be a no-op. GitHub's PR page shows
`mergeStateStatus: BEHIND` (and a UI banner: "This branch is out-of-date
with the base branch") for exactly this case — that banner's "Update
branch" button creates a **merge commit**, which this repo doesn't want
(see the "no merge commits" convention above). Treat that banner as the
trigger for this step's rebase, never the button.

**3. Rebase onto main if needed.**

```bash
git checkout <branch>
git rebase origin/main
```

- On real conflicts: stop, report the conflicting files, and let the
  user decide how to resolve — don't guess at intent.
- If it instead fails with `error: Your local changes to the following
  files would be overwritten by merge` on files you never touched by
  hand, and this repeats on retry (including with
  `git -c core.hooksPath=/dev/null rebase origin/main`, which rules out
  a `pre-commit`/lint-staged interaction) — that's not a real conflict,
  it's an environment/git quirk (observed with multiple small commits
  touching the same lines, executable-bit files involved). Don't fight
  it; squash onto the new base instead, which sidesteps the buggy
  multi-commit replay entirely:

  ```bash
  git branch backup/<branch>-presquash   # safety net, delete once confirmed fine
  git reset --soft origin/main
  git status   # any file that only exists on origin/main now shows as "deleted" — restore those:
  git restore --source=origin/main --staged --worktree -- <those files>
  git commit -m "..."                    # one clean commit with everything <branch> was adding
  ```

  This is squashing your own feature branch's commits, not the PR merge
  itself (still a plain `git merge --ff-only` in step 4) — the "no
  squash" convention is about the merge method into `main`, not about
  how many commits a branch carries getting there.

- On success (rebase or squash-reset both rewrite commits), update the
  PR branch:

```bash
git push --force-with-lease origin <branch>
```

Then re-check CI on the rebased commit before continuing (step 1).

**4. Fast-forward `main` and push.**

```bash
git checkout main
git merge --ff-only origin/main
git merge --ff-only <branch>
git push origin main
```

If either `merge --ff-only` refuses (non-fast-forward), stop — that means
`main` or the branch moved unexpectedly since step 2. Re-fetch and
re-evaluate rather than forcing.

**5. Verify.**

```bash
gh pr view <N> --json state,mergeCommit
```

Confirm `state` is `MERGED`.

## After merging

Merging to `main` does **not** deploy. This repo deploys via `make
deploy` (Kamal, targets the real production server) — that command is
human-operator-only; never run it yourself. If the change is
user-visible, mention that it still needs a deploy to take effect on
`checkmeup.net`.

## Cleanup (optional)

The remote branch can be deleted after a successful merge if the user
wants:

```bash
git push origin --delete <branch>
git branch -d <branch>
```

Only do this if asked — leaving merged branches around is harmless.
