"""Release authorization tests use synthetic data and disposable signing keys."""

import importlib.util
import hashlib
import shutil
import base64
import json
import os
import re
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPTS = Path(__file__).resolve().parent
ROOT = SCRIPTS.parent


def workflow_script(name):
    workflow = (SCRIPTS.parent / ".github/workflows/cd.yml").read_text()
    step = workflow.split("      - name: " + name + "\n", 1)[1].split("      - name:", 1)[0]
    body = step.split("        run: |\n", 1)[1]
    return "\n".join(line[10:] for line in body.splitlines() if line.startswith("          "))


class SecurityTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        spec = importlib.util.spec_from_file_location("release_gate", SCRIPTS / "release-gate.py")
        cls.gate = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.gate)

    def test_security_severities_and_unknown_fail_closed(self):
        for level, expected in (("low", 0), ("medium", 1), ("moderate", 1),
                                ("high", 1), ("critical", 1), (None, 1)):
            with self.subTest(level=level):
                self.assertEqual(self.gate.blocking_count("code", [
                    {"rule": {"security_severity_level": level}}
                ]), expected)
                self.assertEqual(self.gate.blocking_count("dependencies", [
                    {"security_advisory": {"severity": level}}
                ]), expected)

    def test_advisories_and_secrets(self):
        self.assertEqual(self.gate.blocking_count("advisories", [
            {"state": "published", "severity": "critical"},
            {"state": "draft", "severity": "medium"},
            {"state": "triage", "severity": None},
            {"state": "closed", "severity": "critical"},
            {"state": "withdrawn", "severity": "critical"},
        ]), 3)
        self.assertEqual(self.gate.blocking_count("secrets", [{"secret": "not-logged"}]), 1)

    def test_publishing_jobs_require_both_release_gates(self):
        workflow = (SCRIPTS.parent / ".github/workflows/cd.yml").read_text()
        jobs = {}
        current = None
        for line in workflow.splitlines():
            job = re.fullmatch(r"  ([\w-]+):", line)
            if job:
                current = job[1]
                jobs[current] = []
            elif current:
                jobs[current].append(line)
        publishers = []
        for name, lines in jobs.items():
            active = [line for line in lines if not line.lstrip().startswith("#")]
            if not any(re.search(r"\bgh release (?:create|upload|edit)\b", line) for line in active):
                continue
            publishers.append(name)
            needs = next((line for line in active if line.startswith("    needs:")), "")
            dependencies = set(re.findall(r"[\w-]+", needs.partition(":")[2]))
            self.assertTrue({"verify-release", "security-gate"} <= dependencies, name)
        self.assertEqual(publishers, ["build-and-deploy"])
        build_needs = next(line for line in jobs["build-pybindings"] if line.startswith("    needs:"))
        self.assertTrue({"verify-release", "security-gate"} <= set(re.findall(r"[\w-]+", build_needs.partition(":")[2])))

    def test_goreleaser_publication_is_disabled(self):
        config = (ROOT / ".goreleaser.yml").read_text()
        release = re.search(r"(?m)^release:\n((?:[ \t].*\n|\n)*)", config)
        self.assertIsNotNone(release, "GoReleaser must delegate publication to gh")
        fields = [line.split("#", 1)[0].strip() for line in release[1].splitlines()]
        self.assertIn("disable: true", fields)

    def test_unreadable_security_source_is_not_clean(self):
        from unittest.mock import patch
        with patch.object(self.gate.subprocess, "run", side_effect=subprocess.CalledProcessError(
                1, "gh", stderr="private-response-sentinel")):
            with self.assertRaisesRegex(RuntimeError, "unreadable") as caught:
                self.gate.fetch("fixture")
        self.assertNotIn("private-response-sentinel", str(caught.exception))

    def test_paginated_sources_are_complete_and_malformed_sources_hold(self):
        from unittest.mock import patch
        result = subprocess.CompletedProcess([], 0, stdout='[[{"number":1}],[{"number":2}]]')
        with patch.object(self.gate.subprocess, "run", return_value=result):
            self.assertEqual(len(self.gate.fetch("fixture")), 2)
        for body in ('{}', '[{}]', '[[null]]', 'not JSON'):
            with self.subTest(body=body):
                result.stdout = body
                with patch.object(self.gate.subprocess, "run", return_value=result):
                    with self.assertRaises(RuntimeError):
                        self.gate.fetch("fixture")

    def test_provenance_requires_matching_merged_release_and_milestone(self):
        good = {"merged_at": "2026-11-17T00:00:00Z", "merge_commit_sha": "a" * 40,
                "base": {"ref": "master", "repo": {"full_name": "example/provider"}},
                "head": {"ref": "release/v1.2.3", "repo": {"full_name": "example/provider"}},
                "milestone": {"title": "v1.2.3"}}
        self.assertTrue(self.gate.matches_release(good, "example/provider", "a" * 40, "v1.2.3"))
        for field, value in (("merged_at", None), ("merge_commit_sha", "b" * 40),
                             ("head", {"ref": "feature/other", "repo": {"full_name": "example/provider"}}),
                             ("milestone", None)):
            with self.subTest(field=field):
                self.assertFalse(self.gate.matches_release({**good, field: value},
                                                          "example/provider", "a" * 40, "v1.2.3"))


class SigningTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(prefix="tf-sign-", dir="/tmp")
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.home = self.root / "gnupg"
        self.home.mkdir(mode=0o700)
        self.env = {**os.environ, "GNUPGHOME": str(self.home)}
        self.command("git", "init", "-q")
        self.command("git", "config", "user.name", "Fixture")
        self.command("git", "config", "user.email", "fixture@example.invalid")
        self.command("git", "commit", "--allow-empty", "-qm", "fixture")
        (self.root / ".github").mkdir()
        (self.root / ".github/signing-maintainers").write_text("fixture\n")
        self.bin = self.root / "bin"
        self.bin.mkdir()
        curl = self.bin / "curl"
        curl.write_text("#!/bin/sh\ncase \"$*\" in *gpg_keys*) cat \"$FIXTURE_GPG_INDEX\" ;; *ssh_signing_keys*) cat \"$FIXTURE_SSH_SIGNING\" ;; *.gpg*) [ \"${FIXTURE_GPG_HTTP_FAIL:-0}\" = 0 ] || exit 22; cat \"$FIXTURE_GPG\" ;; *.keys*) cat \"$FIXTURE_SSH\" ;; *) exit 1 ;; esac\n")
        curl.chmod(0o755)
        self.gpg = self.root / "public.gpg"
        self.gpg_index = self.root / "gpg-keys.json"
        self.gpg_index.write_text("[]")
        self.ssh = self.root / "public.keys"
        self.ssh_signing = self.root / "signing-keys.json"
        self.ssh_signing.write_text("[]")
        self.gpg.touch()
        self.ssh.touch()
        self.env.update(PATH=str(self.bin) + os.pathsep + os.environ["PATH"],
                        FIXTURE_GPG=str(self.gpg), FIXTURE_SSH=str(self.ssh),
                        FIXTURE_GPG_INDEX=str(self.gpg_index),
                        FIXTURE_SSH_SIGNING=str(self.ssh_signing))

    def enroll_ssh(self, key):
        self.ssh_signing.write_text(json.dumps([{"key": key.strip()}]))

    def command(self, *args):
        return subprocess.run(args, cwd=self.root, env=self.env, check=True,
                              capture_output=True, text=True)

    def verify(self, enforce):
        return subprocess.run(["bash", str(SCRIPTS / "verify-release-tag.sh"), "v1.2.3"],
                              cwd=self.root, env={**self.env, "ENFORCE_SIGNED_TAGS": enforce},
                              capture_output=True, text=True)

    def test_unsigned_tag_fails_when_enforced_and_warns_by_default(self):
        self.command("git", "tag", "-a", "v1.2.3", "-m", "fixture")
        self.assertNotEqual(self.verify("true").returncode, 0)
        result = self.verify("false")
        self.assertEqual(result.returncode, 0)
        self.assertIn("warning", result.stdout.lower())

    def test_enrolled_gpg_signer_passes(self):
        self.command("gpg", "--batch", "--pinentry-mode", "loopback", "--passphrase", "",
                     "--quick-generate-key", "Fixture <fixture@example.invalid>", "ed25519", "sign", "1d")
        key = self.command("gpg", "--with-colons", "--list-secret-keys").stdout
        fingerprint = next(line.split(":")[9] for line in key.splitlines() if line.startswith("fpr:"))
        self.gpg.write_text(self.command("gpg", "--armor", "--export", fingerprint).stdout)
        self.gpg_index.write_text('[{"id": 1}]')
        self.command("git", "-c", "user.signingkey=" + fingerprint, "tag", "-s", "v1.2.3", "-m", "fixture")
        result = self.verify("true")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_enrolled_ssh_signer_passes(self):
        private = self.root / "fixture-ssh"
        self.command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(private))
        self.enroll_ssh(private.with_suffix(".pub").read_text())
        self.command("git", "-c", "gpg.format=ssh", "-c", "user.signingkey=" + str(private),
                     "tag", "-s", "v1.2.3", "-m", "fixture")
        result = self.verify("true")
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_unenrolled_signer_is_rejected(self):
        private = self.root / "fixture-ssh"
        self.command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(private))
        other = self.root / "other-ssh"
        self.command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(other))
        self.enroll_ssh(other.with_suffix(".pub").read_text())
        self.command("git", "-c", "gpg.format=ssh", "-c", "user.signingkey=" + str(private),
                     "tag", "-s", "v1.2.3", "-m", "fixture")
        self.assertNotEqual(self.verify("true").returncode, 0)

    def test_authentication_key_is_not_enrolled_for_signing(self):
        private = self.root / "auth-only"
        self.command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(private))
        self.ssh.write_text(private.with_suffix(".pub").read_text())
        self.command("git", "-c", "gpg.format=ssh", "-c", "user.signingkey=" + str(private),
                     "tag", "-s", "v1.2.3", "-m", "fixture")
        self.assertNotEqual(self.verify("true").returncode, 0)

    def test_ssh_only_enrollment_handles_missing_gpg_export(self):
        private = self.root / "ssh-only"
        self.command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(private))
        self.enroll_ssh(private.with_suffix(".pub").read_text())
        self.command("git", "-c", "gpg.format=ssh", "-c", "user.signingkey=" + str(private),
                     "tag", "-s", "v1.2.3", "-m", "fixture")
        for failure, body in (("1", ""), ("0", "This user has no GPG keys.\n")):
            with self.subTest(gpg_export_failure=failure):
                self.env["FIXTURE_GPG_HTTP_FAIL"] = failure
                self.gpg.write_text(body)
                result = self.verify("true")
                self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_hardware_ssh_key_does_not_block_enrolled_gpg_signer(self):
        kind = b"sk-ssh-ed25519@openssh.com"
        fields = (kind, b"x" * 32, b"ssh:")
        blob = b"".join(len(field).to_bytes(4, "big") + field for field in fields)
        public = kind.decode() + " " + base64.b64encode(blob).decode()
        self.enroll_ssh(public)
        self.ssh.write_text(public + "\n")
        self.test_enrolled_gpg_signer_passes()

    def test_signing_key_source_failure_holds(self):
        self.command("git", "tag", "-a", "v1.2.3", "-m", "fixture")
        (self.bin / "curl").write_text("#!/bin/sh\nexit 22\n")
        self.assertNotEqual(self.verify("true").returncode, 0)


    def test_inline_registry_import_signs_and_cleans_up_without_secret_output(self):
        passphrase = "disposable-passphrase-sentinel"
        self.command("gpg", "--batch", "--pinentry-mode", "loopback", "--passphrase", passphrase,
                     "--quick-generate-key", "Fixture <fixture@example.invalid>", "ed25519", "sign", "1d")
        listing = self.command("gpg", "--with-colons", "--list-secret-keys").stdout
        fingerprint = next(line.split(":")[9] for line in listing.splitlines() if line.startswith("fpr:"))
        secret = self.command("gpg", "--batch", "--pinentry-mode", "loopback", "--passphrase", passphrase,
                              "--armor", "--export-secret-keys", fingerprint).stdout
        envfile = self.root / "github-env"
        output = self.root / "github-output"
        env = {**self.env, "RUNNER_TEMP": str(self.root), "GITHUB_ENV": str(envfile),
               "GITHUB_OUTPUT": str(output), "REGISTRY_GPG_PRIVATE_KEY": secret,
               "REGISTRY_GPG_PASSPHRASE": passphrase, "EXPECTED_FINGERPRINT": fingerprint}
        imported = subprocess.run(["bash", "-c", workflow_script("Import GPG key")],
                                  cwd=self.root, env=env, capture_output=True, text=True)
        self.assertEqual(imported.returncode, 0, imported.stdout + imported.stderr)
        self.assertNotIn(passphrase, imported.stdout + imported.stderr)
        self.assertNotIn("PRIVATE KEY BLOCK", imported.stdout + imported.stderr)
        imported_env = dict(line.split("=", 1) for line in envfile.read_text().splitlines())
        keyhome = Path(imported_env["GNUPGHOME"])
        password_file = Path(imported_env["GPG_PASSPHRASE_FILE"])
        self.assertEqual(keyhome.stat().st_mode & 0o777, 0o700)
        self.assertEqual(password_file.stat().st_mode & 0o777, 0o600)
        self.assertEqual(output.read_text(), "fingerprint=" + fingerprint + "\n")
        checksums = self.root / "fixture_SHA256SUMS"
        checksums.write_text("0" * 64 + "  fixture.zip\n")
        signed = subprocess.run(["gpg", "--batch", "--pinentry-mode", "loopback", "--passphrase-file",
                                 str(password_file), "--local-user", fingerprint, "--detach-sign", str(checksums)],
                                env={**self.env, **imported_env}, capture_output=True)
        self.assertEqual(signed.returncode, 0)
        self.command("gpg", "--verify", str(checksums) + ".sig", str(checksums))
        cleanup = subprocess.run(["bash", "-c", workflow_script("Remove signing material")],
                                 env={**self.env, **imported_env, "RUNNER_TEMP": str(self.root)}, capture_output=True)
        self.assertEqual(cleanup.returncode, 0)
        self.assertFalse(keyhome.exists())

    def test_workflow_actions_are_immutable(self):
        for path in (SCRIPTS.parent / ".github/workflows").glob("*.yml"):
            for action in re.findall(r"uses:\s*([^\s]+)", path.read_text()):
                self.assertRegex(action, r"^[^@]+@[0-9a-f]{40}$", path.name)

    def test_inline_import_failure_is_generic_and_still_cleans_up(self):
        envfile = self.root / "github-env"
        sentinel = "invalid-fixture-signing-material"
        env = {**self.env, "RUNNER_TEMP": str(self.root), "GITHUB_ENV": str(envfile),
               "GITHUB_OUTPUT": str(self.root / "output"), "REGISTRY_GPG_PRIVATE_KEY": sentinel,
               "REGISTRY_GPG_PASSPHRASE": "", "EXPECTED_FINGERPRINT": ""}
        result = subprocess.run(["bash", "-c", workflow_script("Import GPG key")],
                                cwd=self.root, env=env, capture_output=True, text=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertNotIn(sentinel, result.stdout + result.stderr)
        imported_env = dict(line.split("=", 1) for line in envfile.read_text().splitlines())
        subprocess.run(["bash", "-c", workflow_script("Remove signing material")],
                       env={**self.env, **imported_env, "RUNNER_TEMP": str(self.root)}, check=True)
        self.assertFalse(Path(imported_env["GNUPGHOME"]).exists())


class PublicationTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.calls = self.root / "calls.jsonl"
        bindir = self.root / "bin"
        bindir.mkdir()
        gh = bindir / "gh"
        gh.write_text('''#!/usr/bin/env python3
import json, os, sys
with open(os.environ["CALL_LOG"], "a") as log:
    log.write(json.dumps(sys.argv[1:]) + "\\n")
if os.environ.get("FAIL_CREATE") == "1" and sys.argv[1:3] == ["release", "create"]:
    sys.exit(7)
''')
        gh.chmod(0o755)
        dist = self.root / "dist"
        dist.mkdir()
        self.artifacts = {"a.zip", "b.zip", "c.zip", "d.zip", "e.zip", "fixture_SHA256SUMS", "fixture_SHA256SUMS.sig"}
        for name in self.artifacts:
            (dist / name).touch()
        (self.root / "release-notes.md").write_text("Fixture notes\n")
        self.env = {**os.environ, "RELEASE_TAG": "v1.2.3", "RUNNER_TEMP": str(self.root),
                    "CALL_LOG": str(self.calls), "PATH": str(bindir) + os.pathsep + os.environ["PATH"]}

    def publish(self, **env):
        result = subprocess.run(["bash", "-c", workflow_script("Publish GitHub release")],
                                cwd=self.root, env={**self.env, **env}, capture_output=True, text=True)
        calls = [json.loads(line) for line in self.calls.read_text().splitlines()]
        return result, calls

    def test_draft_upload_precedes_final_publication(self):
        result, calls = self.publish()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(len(calls), 2)
        create, edit = calls
        self.assertEqual(create[:3], ["release", "create", "v1.2.3"])
        self.assertIn("--draft", create)
        self.assertTrue({"dist/" + name for name in self.artifacts} <= set(create))
        self.assertEqual(create[create.index("--notes-file") + 1], str(self.root / "release-notes.md"))
        self.assertEqual(edit, ["release", "edit", "v1.2.3", "--draft=false"])

    def test_failed_upload_does_not_publish(self):
        result, calls = self.publish(FAIL_CREATE="1")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual([call[:3] for call in calls], [["release", "create", "v1.2.3"]])


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
