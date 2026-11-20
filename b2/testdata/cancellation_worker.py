######################################################################
#
# File: b2/testdata/cancellation_worker.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

"""Offline process-tree fixture, runnable as plain Python or frozen with PyInstaller."""

import json
import os
import socket
import subprocess
import sys
import time
from pathlib import Path


def connect():
    host, port = os.environ['B2_TEST_CANCEL_ADDRESS'].rsplit(':', 1)
    connection = socket.create_connection((host, int(port)), timeout=5)
    connection.settimeout(None)
    return connection


if '--grandchild' in sys.argv:
    pids = {'parent': int(os.environ['B2_TEST_LAUNCHER_PID']), 'grandchild': os.getpid()}
    ready = json.dumps(pids).encode() + b'\n'
    sys.stdout.buffer.write(ready)
    sys.stdout.buffer.flush()
    with connect() as connection:
        connection.sendall(ready)
        # Survive launcher-only termination; no parent pipe controls this lifetime.
        time.sleep(300)
else:
    sys.stdin.buffer.read()
    if os.environ.get('B2_TEST_PLAIN_LAUNCHER') == '1':
        environment = dict(os.environ, B2_TEST_LAUNCHER_PID=str(os.getpid()))
        worker = subprocess.Popen(
            [sys.executable, __file__, '--grandchild'],
            stdin=subprocess.DEVNULL,
            env=environment,
        )
        # Leave a cleanup record even if the worker cannot send its ready message.
        Path(os.environ['B2_TEST_CANCEL_PIDS']).write_text(
            json.dumps({'parent': os.getpid(), 'grandchild': worker.pid})
        )
        worker.wait()
    else:
        with connect() as connection:
            connection.sendall(b'started\n')
            connection.recv(1)
