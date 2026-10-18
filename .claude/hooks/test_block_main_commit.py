"""Tests for block_main_commit.py: run with `python3 -m unittest discover -s .claude/hooks`."""

import os
import subprocess  # nosec B404 - fixed argv below, never shell=True
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from block_main_commit import commit_on_main  # noqa: E402


def _repo(root: str, name: str, branch: str) -> str:
    path = os.path.join(root, name)
    for argv in (["init", "-q", "-b", branch, path], ["-C", path, "commit", "-q", "--allow-empty", "-m", "init"]):
        subprocess.run(  # nosec B603 B607 - fixed argv, no shell
            ["git", "-c", "user.name=t", "-c", "user.email=t@t", *argv], check=True, capture_output=True
        )
    return path


class CommitOnMainTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls._tmp = tempfile.TemporaryDirectory()
        cls.main = _repo(cls._tmp.name, "on-main", "main")
        cls.feature = _repo(cls._tmp.name, "on-feature", "feature")

    @classmethod
    def tearDownClass(cls) -> None:
        cls._tmp.cleanup()

    def test_commit_on_main_is_blocked(self) -> None:
        self.assertTrue(commit_on_main("git commit -m x", self.main))

    def test_commit_on_a_feature_branch_is_allowed(self) -> None:
        self.assertFalse(commit_on_main("git commit -m x", self.feature))

    def test_a_global_option_before_commit_still_counts(self) -> None:
        self.assertTrue(commit_on_main("git -c core.hooksPath=/dev/null commit -m x", self.main))

    def test_creating_a_branch_earlier_in_the_command_counts(self) -> None:
        self.assertFalse(commit_on_main("git switch -c fix/x && git commit -m x", self.main))
        self.assertFalse(commit_on_main("git checkout -b fix/x; git commit -m x", self.main))

    def test_switching_to_main_earlier_in_the_command_counts(self) -> None:
        self.assertTrue(commit_on_main("git switch main && git commit -m x", self.feature))

    def test_cd_into_another_repo_reads_that_repo(self) -> None:
        self.assertFalse(commit_on_main(f"cd {self.feature} && git commit -m x", self.main))
        self.assertTrue(commit_on_main(f"cd {self.main} && git commit -m x", self.feature))

    def test_dash_c_reads_that_repo(self) -> None:
        self.assertFalse(commit_on_main(f"git -C {self.feature} commit -m x", self.main))
        self.assertTrue(commit_on_main(f"git -C {self.main} commit -m x", self.feature))

    def test_commit_text_in_quotes_or_a_heredoc_is_ignored(self) -> None:
        self.assertFalse(commit_on_main('echo "git commit -m x"', self.main))
        self.assertFalse(commit_on_main("cat > notes.md <<'EOF'\ngit commit -m x\nEOF\necho done", self.main))

    def test_outside_a_repo_is_allowed(self) -> None:
        self.assertFalse(commit_on_main("git commit -m x", self._tmp.name))


if __name__ == "__main__":
    unittest.main()
