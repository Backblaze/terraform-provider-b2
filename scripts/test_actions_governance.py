######################################################################
#
# File: scripts/test_actions_governance.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import base64
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[1]
CHANGELOG = ROOT / 'scripts/actions-changelog.py'
IMPORT_KEY = ROOT / 'scripts/import-release-key.sh'
FIXTURE = '''# Changelog

## [Unreleased]

## [1.2.3] - 2026-11-18

### Fixed
* Preserve [reference][] and literal $HOME text.

## [1.2.2] - 2026-11-17

### Changed
* Older entry.

[reference]: https://example.invalid/reference
'''
NOTES = '\n### Fixed\n* Preserve [reference][] and literal $HOME text.\n[reference]: https://example.invalid/reference'


class ChangelogTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.path = Path(self.tmp.name)
        (self.path / 'CHANGELOG.md').write_text(FIXTURE)

    def run_script(self, *args):
        return subprocess.run(
            [sys.executable, str(CHANGELOG), *args], cwd=self.path,
            text=True, capture_output=True,
            env={**os.environ, 'GITHUB_OUTPUT': str(self.path / 'outputs')},
        )

    def test_extract_release_body_and_outputs(self):
        result = self.run_script('extract', '1.2.3', '--notes-file', 'notes.md')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual((self.path / 'notes.md').read_text(), NOTES)
        output = (self.path / 'outputs').read_text()
        self.assertIn('version=1.2.3\n', output)
        self.assertIn('date=2026-11-18\n', output)
        self.assertIn('status=released\n', output)
        self.assertIn(NOTES + '\n', output)

    def test_missing_version_is_rejected(self):
        result = self.run_script('extract', '9.9.9', '--notes-file', 'notes.md')
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('not found', result.stderr)
        self.assertFalse((self.path / 'notes.md').exists())

    def test_format_and_pr_presence_with_label_opt_out(self):
        event = {'pull_request': {'labels': [], 'head': {'ref': 'ci/example'}}}
        (self.path / 'event.json').write_text(json.dumps(event))
        (self.path / 'files').write_text('b2/client.go\n')
        args = ('validate', '--event-file', 'event.json', '--changed-files', 'files')
        result = self.run_script(*args)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('CHANGELOG.md must be modified', result.stderr)
        event['pull_request']['labels'] = [{'name': '-changelog'}]
        (self.path / 'event.json').write_text(json.dumps(event))
        self.assertEqual(self.run_script(*args).returncode, 0)
        event['pull_request']['labels'] = []
        (self.path / 'event.json').write_text(json.dumps(event))
        (self.path / 'files').write_text('CHANGELOG.md\n')
        self.assertEqual(self.run_script(*args).returncode, 0)

    def test_malformed_duplicate_and_unordered_entries_are_rejected(self):
        for text in (
            FIXTURE.replace('2026-11-18', '2026-99-18'),
            FIXTURE.replace('[1.2.2]', '[1.2.3]'),
            FIXTURE.replace('[1.2.2]', '[2.0.0]'),
            FIXTURE.replace('### Fixed', '### Unknown'),
        ):
            with self.subTest(text=text):
                (self.path / 'CHANGELOG.md').write_text(text)
                result = self.run_script('validate')
                self.assertNotEqual(result.returncode, 0)
                self.assertIn('Invalid changelog', result.stderr)


class SigningTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        # Keep Unix socket paths short on macOS as well as Linux runners.
        cls.tmp = tempfile.TemporaryDirectory(prefix='act-', dir='/tmp')
        cls.path = Path(cls.tmp.name)
        cls.home = cls.path / 'source'
        cls.home.mkdir(mode=0o700)
        cls.passphrase = 'disposable-fixture-passphrase'
        cls.env = {**os.environ, 'GNUPGHOME': str(cls.home)}
        gpg = ['gpg', '--batch', '--pinentry-mode', 'loopback', '--passphrase-fd', '0']
        subprocess.run(
            [*gpg, '--quick-generate-key', 'Workflow Fixture <fixture@example.invalid>',
             'ed25519', 'cert', '0'], input=cls.passphrase, env=cls.env,
            text=True, capture_output=True, check=True,
        )
        listed = subprocess.check_output(
            ['gpg', '--batch', '--with-colons', '--list-secret-keys'], env=cls.env, text=True,
            stderr=subprocess.DEVNULL,
        )
        cls.fingerprint = next(line.split(':')[9] for line in listed.splitlines() if line.startswith('fpr:'))
        subprocess.run(
            [*gpg, '--quick-add-key', cls.fingerprint, 'ed25519', 'sign', '0'],
            input=cls.passphrase, env=cls.env, text=True, capture_output=True, check=True,
        )
        cls.key = subprocess.check_output(
            [*gpg, '--armor', '--export-secret-keys', cls.fingerprint],
            input=cls.passphrase, env=cls.env, text=True, stderr=subprocess.DEVNULL,
        )

    @classmethod
    def tearDownClass(cls):
        subprocess.run(['gpgconf', '--kill', 'gpg-agent'], env=cls.env, capture_output=True)
        cls.tmp.cleanup()

    def test_protected_key_import_and_goreleaser_signing(self):
        for key in (self.key, base64.b64encode(self.key.encode()).decode()):
            with self.subTest(encoding='armored' if key == self.key else 'base64'):
                with tempfile.TemporaryDirectory(prefix='act-', dir='/tmp') as directory:
                    tmp = Path(directory)
                    output, environment = tmp / 'output', tmp / 'environment'
                    env = {**os.environ, 'RUNNER_TEMP': directory, 'GITHUB_OUTPUT': str(output),
                           'GITHUB_ENV': str(environment), 'GPG_PRIVATE_KEY': key,
                           'PASSPHRASE': self.passphrase}
                    result = subprocess.run(['bash', str(IMPORT_KEY)], env=env, text=True, capture_output=True)
                    self.assertEqual(result.returncode, 0, result.stderr)
                    values = dict(line.split('=', 1) for line in environment.read_text().splitlines())
                    try:
                        self.assertEqual(values['GPG_FINGERPRINT'], self.fingerprint)
                        self.assertIn('fingerprint=' + self.fingerprint, output.read_text())
                        env['GNUPGHOME'] = values['GNUPGHOME']
                        artifact, signature = tmp / 'SHA256SUMS', tmp / 'SHA256SUMS.sig'
                        artifact.write_text('fixture checksum contents\n')
                        signed = subprocess.run(
                            ['gpg', '--batch', '--local-user', self.fingerprint,
                             '--output', str(signature), '--detach-sign', str(artifact)],
                            env=env, text=True, capture_output=True,
                        )
                        self.assertEqual(signed.returncode, 0, signed.stderr)
                        verified = subprocess.run(
                            ['gpg', '--batch', '--verify', str(signature), str(artifact)],
                            env=env, capture_output=True,
                        )
                        self.assertEqual(verified.returncode, 0)
                        self.assertNotIn(self.passphrase, result.stdout + result.stderr)
                        self.assertNotIn(self.key, result.stdout + result.stderr)
                        print(f'\nPASS: disposable {"armored" if key == self.key else "base64"} key fingerprint {self.fingerprint}; checksum signature verified')
                    finally:
                        subprocess.run(['gpgconf', '--kill', 'gpg-agent'], env={**env, **values}, capture_output=True)

    def test_invalid_key_is_rejected_without_echoing_it(self):
        with tempfile.TemporaryDirectory(prefix='act-', dir='/tmp') as directory:
            env = {**os.environ, 'RUNNER_TEMP': directory,
                   'GITHUB_OUTPUT': directory + '/output', 'GITHUB_ENV': directory + '/environment',
                   'GPG_PRIVATE_KEY': 'invalid-disposable-fixture-key', 'PASSPHRASE': ''}
            result = subprocess.run(['bash', str(IMPORT_KEY)], env=env, text=True, capture_output=True)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn('Could not import release signing key', result.stderr)
            self.assertNotIn(env['GPG_PRIVATE_KEY'], result.stdout + result.stderr)


if __name__ == '__main__':
    unittest.main(verbosity=2)
