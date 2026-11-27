######################################################################
#
# File: python-bindings/tests/b2_simulator.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################
"""Start a standalone b2-simulator (backblaze-labs/b2-simulator) for tests.

Set B2_SIMULATOR_DIR to a checkout of b2-simulator; Node.js must be on PATH. The simulator listens on
loopback only and accepts only its fixed test credentials, so no real B2 account or network is used.
"""

import json
import os
import subprocess
import urllib.request

# The simulator's fixed test credentials (b2-simulator docs/CONTRACT.md).
KEY_ID = 'sim-test-id'
KEY = 'sim-test-secret'

# Runs serve.mjs and stops it (SIGTERM, its normal shutdown) when its stdin closes, so the simulator also
# exits when the process that started it is killed without cleanup.
_LAUNCHER = """
import { pathToFileURL } from 'node:url';
const [serve, ...args] = process.argv.slice(1);
process.argv = [process.argv[0], serve, ...args];
process.stdin.on('end', () => process.kill(process.pid, 'SIGTERM'));
process.stdin.on('error', () => process.kill(process.pid, 'SIGTERM'));
process.stdin.resume();
await import(pathToFileURL(serve).href);
"""


class B2Simulator:
    """One `serve.mjs --control --profile b2` process."""

    def __init__(self, simulator_dir=None):
        simulator_dir = simulator_dir or os.environ.get('B2_SIMULATOR_DIR')
        if not simulator_dir:
            raise RuntimeError(
                'B2_SIMULATOR_DIR is not set: these tests run against b2-simulator. '
                'Set it to a checkout of backblaze-labs/b2-simulator (v0.6.0 or later).'
            )
        serve = os.path.join(simulator_dir, 'bin', 'simulator', 'serve.mjs')
        if not os.path.isfile(serve):
            raise RuntimeError(f'B2_SIMULATOR_DIR={simulator_dir!r} has no bin/simulator/serve.mjs')
        self._process = subprocess.Popen(
            ['node', '--input-type=module', '-e', _LAUNCHER, serve, '--control', '--profile', 'b2'],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
        )
        # With --control the simulator prints `SIMULATOR-LISTENING http://...`, `SIMULATOR-LISTENING https://...`
        # and `SIMULATOR-CONTROL http://...`. The plain HTTP listener avoids trusting its loopback certificate.
        lines = [self._process.stdout.readline().split() for _ in range(3)]
        urls = {}
        for line in lines:
            if len(line) == 2:
                urls.setdefault(line[0], []).append(line[1])
        api_urls = [url for url in urls.get('SIMULATOR-LISTENING', []) if url.startswith('http://')]
        control_urls = urls.get('SIMULATOR-CONTROL', [])
        if not api_urls or not control_urls:
            self.stop()
            raise RuntimeError(f'b2-simulator did not start: {lines!r}')
        self.api_url = api_urls[0]
        self.control_url = control_urls[0]

    def control(self, path, body=None):
        """POST to the simulator's control plane (reset, clock, faults...)."""
        request = urllib.request.Request(
            self.control_url + path,
            data=json.dumps(body or {}).encode(),
            method='POST',
            headers={'Content-Type': 'application/json'},
        )
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
        with opener.open(request, timeout=30) as response:
            return json.loads(response.read() or b'{}')

    def reset(self):
        """Empty the store and return the virtual clock to real time."""
        self.control('/reset')

    def advance_clock(self, milliseconds):
        self.control('/clock', {'advanceMs': milliseconds})

    def stop(self):
        if self._process.poll() is None:
            self._process.terminate()
            self._process.wait(timeout=10)
        self._process.stdin.close()
        self._process.stdout.close()
