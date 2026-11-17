"""Offline release workflow contract tests; no production keys or API writes."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPTS = Path(__file__).resolve().parent


class ReleaseTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)

    def git(self, *args):
        return subprocess.run(["git", *args], cwd=self.root, check=True,
                              capture_output=True, text=True).stdout.strip()

    def run_script(self, name, *args, **env):
        return subprocess.run(["bash", str(SCRIPTS / name), *args], cwd=self.root,
                              env={**os.environ, **env}, capture_output=True, text=True)

    def test_notes_use_assembled_version(self):
        (self.root / ".changes").mkdir()
        (self.root / ".changes/v1.2.3.md").write_text(
            "## [1.2.3] - 2026-11-17\n\n### Fixed\n* Example.\n")
        result = self.run_script("release-notes.sh", "v1.2.3")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout, "\n### Fixed\n* Example.\n")

    def test_notes_reject_missing_or_mismatched_version(self):
        (self.root / ".changes").mkdir()
        (self.root / ".changes/v1.2.3.md").write_text("## [9.9.9] - 2026-11-17\n")
        for tag in ("v1.2.3", "v1.2.4", "../../outside"):
            with self.subTest(tag=tag):
                self.assertNotEqual(self.run_script("release-notes.sh", tag).returncode, 0)

    def test_changelog_presence_and_label_contract(self):
        self.git("init", "-q")
        self.git("config", "user.name", "Fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        (self.root / "README.md").write_text("Before\n")
        self.git("add", ".")
        self.git("commit", "-qm", "base")
        base = self.git("rev-parse", "HEAD")
        (self.root / "README.md").write_text("After\n")
        self.git("add", ".")
        self.git("commit", "-qm", "change")
        self.assertNotEqual(self.run_script("changelog-check.sh", base).returncode, 0)
        self.assertEqual(self.run_script("changelog-check.sh", base,
                                        PR_LABELS_JSON='["-changelog"]').returncode, 0)
        changes = self.root / ".changes/unreleased"
        changes.mkdir(parents=True)
        (changes / "example.yaml").write_text(
            "kind: Infrastructure\nbody: Example.\ntime: 2026-11-17T00:00:00Z\n")
        self.git("add", ".")
        self.git("commit", "-qm", "fragment")
        self.assertEqual(self.run_script("changelog-check.sh", base).returncode, 0)

    def test_release_branch_required_for_assembled_changelog(self):
        self.git("init", "-q")
        self.git("config", "user.name", "Fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        (self.root / ".changie.yaml").write_text("changesDir: .changes\n")
        (self.root / "CHANGELOG.md").write_text("# Changelog\n")
        (self.root / "provider.go").write_text("package fixture\n")
        self.git("add", ".")
        self.git("commit", "-qm", "base")
        base = self.git("rev-parse", "HEAD")
        (self.root / ".changes").mkdir()
        (self.root / ".changes/v1.2.3.md").write_text("## [1.2.3] - 2026-11-17\n")
        (self.root / "CHANGELOG.md").write_text("# Changelog\n\n## [1.2.3]\n")
        self.git("add", ".")
        self.git("commit", "-qm", "release")
        self.assertNotEqual(self.run_script("changelog-check.sh", base,
                                           PR_HEAD_REF="feature/example").returncode, 0)
        self.assertEqual(self.run_script("changelog-check.sh", base,
                                        PR_HEAD_REF="release/v1.2.3").returncode, 0)
        (self.root / "provider.go").unlink()
        self.git("add", "-u")
        self.git("commit", "-qm", "unexpected source deletion")
        self.assertNotEqual(self.run_script("changelog-check.sh", base,
                                           PR_HEAD_REF="release/v1.2.3").returncode, 0)


if __name__ == "__main__":
    unittest.main(verbosity=2)
