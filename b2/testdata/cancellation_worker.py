######################################################################
#
# File: b2/testdata/cancellation_worker.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

"""Offline worker fixture; freeze with PyInstaller --onefile for the process-tree test."""

import os
import socket
import sys

sys.stdin.buffer.read()
host, port = os.environ['B2_TEST_CANCEL_ADDRESS'].rsplit(':', 1)
with socket.create_connection((host, int(port)), timeout=5) as connection:
    connection.settimeout(None)
    connection.sendall(b'started\n')
    connection.recv(1)
