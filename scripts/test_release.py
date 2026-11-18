"""Offline release workflow contract tests; no production keys or API writes."""

import os
import re
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


    def test_release_branch_rejects_source_change_without_changelog_edit(self):
        self.git("init", "-q")
        self.git("config", "user.name", "Fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        (self.root / ".changie.yaml").write_text("changesDir: .changes\n")
        (self.root / "provider.go").write_text("package fixture\n")
        self.git("add", ".")
        self.git("commit", "-qm", "base")
        base = self.git("rev-parse", "HEAD")
        (self.root / "provider.go").write_text("package changed\n")
        fragments = self.root / ".changes/unreleased"
        fragments.mkdir(parents=True)
        (fragments / "fix.yaml").write_text("kind: Fixed\nbody: Example.\n")
        self.git("add", ".")
        self.git("commit", "-qm", "source and fragment")
        result = self.run_script("changelog-check.sh", base, PR_HEAD_REF="release/v1.2.3")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Release PRs may change only", result.stderr)
        result = self.run_script("changelog-check.sh", base, PR_HEAD_REF="release/v1.2.3",
                                 PR_LABELS_JSON='["-changelog"]')
        self.assertNotEqual(result.returncode, 0)

    def test_migrated_comparison_links_are_generated(self):
        repo = SCRIPTS.parent
        header = (repo / ".changes/header.tpl.md").read_text()
        # Derive the required base from versioned-note inputs, not the header
        # being checked, so a later release needs only allowed version-path edits.
        versions = []
        for path in (repo / ".changes").glob("v*.md"):
            match = re.fullmatch(r"v(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?", path.stem)
            if match:
                prerelease = tuple((0, int(part)) if part.isdigit() else (1, part)
                                   for part in (match[4] or "").split("."))
                versions.append(((int(match[1]), int(match[2]), int(match[3]), match[4] is None, prerelease), path.stem))
        latest = max(versions)[1]
        unreleased = f"[Unreleased]: https://github.com/Backblaze/terraform-provider-b2/compare/{latest}...HEAD"
        self.assertIn(unreleased, header)
        released = "[0.15.0]: https://github.com/Backblaze/terraform-provider-b2/compare/v0.14.1...v0.15.0"
        self.assertIn(unreleased, (repo / "CHANGELOG.md").read_text())
        self.assertIn(released, (repo / ".changes/v0.15.0.md").read_text())
        self.assertIn(released, (repo / "CHANGELOG.md").read_text())

    def test_release_notes_step_does_not_dirty_checkout(self):
        (self.root / ".changes").mkdir()
        (self.root / ".changes/v1.2.3.md").write_text(
            "## [1.2.3] - 2026-11-17\n\n### Fixed\n* Example.\n")
        (self.root / "scripts").symlink_to(SCRIPTS, target_is_directory=True)
        runner_temp = self.root / "runner-temp"
        runner_temp.mkdir()
        workflow = (SCRIPTS.parent / ".github/workflows/cd.yml").read_text()
        step = workflow.split("      - name: Read the Changelog\n", 1)[1].split("      - name:", 1)[0]
        command = next(line.split("run: ", 1)[1] for line in step.splitlines()
                       if line.startswith("        run: "))
        result = subprocess.run(["bash", "-c", command], cwd=self.root,
                                env={**os.environ, "RELEASE_TAG": "v1.2.3", "RUNNER_TEMP": str(runner_temp)},
                                capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse((self.root / "release-notes.md").exists())
        self.assertEqual((runner_temp / "release-notes.md").read_text(), "\n### Fixed\n* Example.\n")
        self.assertIn('--notes-file "$RUNNER_TEMP/release-notes.md"', workflow)

    def test_changelog_workflow_preserves_dependabot_exemption(self):
        workflow = (SCRIPTS.parent / ".github/workflows/changelog.yml").read_text()
        job = workflow.split("  changelog:\n", 1)[1].split("    steps:\n", 1)[0]
        condition = next((line for line in job.splitlines() if line.startswith("    if:")), "")
        self.assertIn("github.actor != 'dependabot[bot]'", condition)
        self.assertIn("github.event.pull_request.user.login != 'dependabot[bot]'", condition)


class ReleaseHealthTests(unittest.TestCase):
    run_script = ReleaseTests.run_script

    def setUp(self):
        ReleaseTests.setUp(self)
        (self.root / "CHANGELOG.md").write_text("# Fixture changelog\n")
        changes = self.root / ".changes/unreleased"
        changes.mkdir(parents=True)
        (changes / "fixture.yaml").write_text("kind: Fixed\nbody: Fixture.\n")
        self.bin = self.root / "bin"
        self.bin.mkdir()
        (self.bin / "go").write_text('''#!/bin/sh
case "$*" in
  *"merge --dry-run") check=merge ;;
  *"batch auto --dry-run") check=batch ;;
  "vet ./...") check=vet ;;
  *) exit 90 ;;
esac
echo "$check" >> "$CHECK_LOG"
[ "$FAIL_CHECK" != "$check" ] || exit 7
if [ "$check" = merge ]; then
  cat CHANGELOG.md
  [ "$FAIL_CHECK" != mismatch ] || echo 'unexpected generated content'
fi
''')
        (self.bin / "make").write_text('''#!/bin/sh
[ "$*" = test ] || exit 90
if printenv TF_ACC >/dev/null; then exit 91; fi
echo make >> "$CHECK_LOG"
[ "$FAIL_CHECK" != make ] || exit 7
''')
        (self.bin / "git").write_text('''#!/bin/sh
[ "$*" = 'diff --check' ] || exit 90
echo git >> "$CHECK_LOG"
[ "$FAIL_CHECK" != git ] || exit 7
''')
        for path in self.bin.iterdir():
            path.chmod(0o755)

    def health(self, failure="", *args):
        log = self.root / "checks"
        log.unlink(missing_ok=True)
        result = self.run_script("release-health.sh", *args, PATH=str(self.bin) + os.pathsep + os.environ["PATH"],
                                 CHECK_LOG=str(log), FAIL_CHECK=failure, TF_ACC="1")
        return result, log.read_text().splitlines()

    def test_health_success_runs_checks_and_clears_acceptance_flag(self):
        result, checks = self.health()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(set(checks), {"merge", "batch", "make", "vet", "git"})
        self.assertIn("Release health: PASS", result.stdout)
        target = subprocess.run(
            ["make", "-f", str(SCRIPTS.parent / "GNUmakefile"), "--dry-run", "release-health"],
            cwd=self.root, capture_output=True, text=True, check=True,
        ).stdout
        self.assertIn("unittest discover -s scripts", target)
        self.assertIn("bash scripts/release-health.sh", target)

    def test_health_failure_never_reports_pass(self):
        for failure in ("merge", "mismatch", "batch", "make", "vet", "git"):
            with self.subTest(failure=failure):
                result, _ = self.health(failure)
                self.assertNotEqual(result.returncode, 0)
                self.assertNotIn("Release health: PASS", result.stdout)

    def test_health_skips_automatic_batch_when_no_fragments_remain(self):
        (self.root / ".changes/unreleased/fixture.yaml").unlink()
        result, checks = self.health("batch")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn("batch", checks)

    def test_changelog_only_runs_generation_checks(self):
        result, checks = self.health("", "--changelog-only")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(set(checks), {"merge", "batch"})
        self.assertNotIn("Release health: PASS", result.stdout)
        workflow = (SCRIPTS.parent / ".github/workflows/changelog.yml").read_text()
        self.assertIn("run: bash scripts/release-health.sh --changelog-only", workflow)


if __name__ == "__main__":
    unittest.main(verbosity=2)
