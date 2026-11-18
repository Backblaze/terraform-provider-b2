"""Offline packaging smoke: real provider builds, fixture bindings, disposable key."""

import hashlib
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[1]


class SnapshotTests(unittest.TestCase):
    def test_five_archives_and_signed_checksums(self):
        goreleaser = shutil.which("goreleaser")
        self.assertIsNotNone(goreleaser, "Install GoReleaser before running release tests")
        with tempfile.TemporaryDirectory(prefix="tf-snapshot-", dir="/tmp") as directory:
            scratch = Path(directory)
            repo = scratch / "repo"
            shutil.copytree(ROOT, repo, ignore=shutil.ignore_patterns(
                ".git", "dist", "__pycache__", "py-terraform-provider-b2"))
            home = scratch / "gpg"
            home.mkdir(mode=0o700)
            password = scratch / "passphrase"
            password.write_text("disposable-snapshot-fixture")
            password.chmod(0o600)
            env = {**os.environ, "GNUPGHOME": str(home),
                   "GPG_PASSPHRASE_FILE": str(password)}

            def run(*args):
                result = subprocess.run(args, cwd=repo, env=env, capture_output=True,
                                        text=True, timeout=600)
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
                return result.stdout

            try:
                run("gpg", "--batch", "--pinentry-mode", "loopback", "--passphrase-file",
                    str(password), "--quick-generate-key", "Snapshot Fixture <fixture@example.invalid>",
                    "ed25519", "sign", "1d")
                keys = run("gpg", "--with-colons", "--list-secret-keys")
                env["GPG_FINGERPRINT"] = next(line.split(":")[9] for line in keys.splitlines() if line.startswith("fpr:"))
                targets = {"linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64", "windows_amd64"}
                bindings = repo / "python-bindings/dist"
                bindings.mkdir(parents=True, exist_ok=True)
                for target in targets:
                    (bindings / ("py-terraform-provider-b2-" + target.replace("_", "-"))).write_bytes(
                        b"OFFLINE PACKAGING FIXTURE, NOT EXECUTABLE\n" + target.encode())
                run("git", "init", "-q")
                run("git", "config", "user.name", "Snapshot Fixture")
                run("git", "config", "user.email", "fixture@example.invalid")
                run("git", "add", ".")
                run("git", "-c", "commit.gpgsign=false", "commit", "-qm", "snapshot fixture")
                run(goreleaser, "release", "--snapshot", "--clean", "--skip=publish", "-p", "1")
                archives = list((repo / "dist").glob("*.zip"))
                self.assertEqual(len(archives), 5)
                for target in targets:
                    self.assertEqual(sum(path.name.endswith("_" + target + ".zip") for path in archives), 1)
                sums = list((repo / "dist").glob("*_SHA256SUMS"))
                self.assertEqual(len(sums), 1)
                checksums = dict(reversed(line.split()) for line in sums[0].read_text().splitlines())
                self.assertEqual(set(checksums), {path.name for path in archives})
                for archive in archives:
                    self.assertEqual(hashlib.sha256(archive.read_bytes()).hexdigest(), checksums[archive.name])
                signature = Path(str(sums[0]) + ".sig")
                self.assertTrue(signature.is_file())
                verified = run("gpg", "--batch", "--status-fd", "1", "--verify", str(signature), str(sums[0]))
                self.assertIn("[GNUPG:] VALIDSIG " + env["GPG_FINGERPRINT"], verified)
            finally:
                subprocess.run(["gpgconf", "--kill", "gpg-agent"], env=env, capture_output=True)


if __name__ == "__main__":
    unittest.main(verbosity=2)
