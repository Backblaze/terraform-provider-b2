#!/usr/bin/env python3
######################################################################
#
# File: b2/testdata/retention_simulator.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################
"""Test helper server for the Terraform retention acceptance test.

It serves the provider's Bucket operations and the test's file hooks from a
b2sdk B2Api connected to a standalone b2-simulator (backblaze-labs/b2-simulator,
v0.6.0 or later), which it starts from B2_SIMULATOR_DIR. The test creates a file
with explicit retention and compares its server-side mode and timestamp before
and after changing the bucket default. After the assertions, cleanup advances
the simulator's virtual clock past the file's retention to allow deletion.
"""

import json
import os
import sys
import time
import urllib.error
import urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer
from types import SimpleNamespace

from b2sdk.v3 import (
    B2Api,
    DummyCache,
    FileRetentionSetting,
    InMemoryAccountInfo,
    RetentionMode,
)

_BINDINGS = os.path.join(os.path.dirname(__file__), "..", "..", "python-bindings")
sys.path.insert(0, _BINDINGS)
sys.path.insert(0, os.path.join(_BINDINGS, "tests"))
from b2_simulator import KEY, KEY_ID, B2Simulator
from b2_terraform.provider_tool import Bucket, ProviderTool


def serve():
    simulator = B2Simulator()
    tool = ProviderTool()
    api = B2Api(account_info=InMemoryAccountInfo(), cache=DummyCache())
    api.authorize_account(KEY_ID, KEY, realm=simulator.api_url)
    tool.__dict__["api"] = api
    files = {}
    # How far the simulator's virtual clock is ahead of real time.
    clock_offset_ms = [0]

    def virtual_now_ms():
        return int(time.time() * 1000) + clock_offset_ms[0]

    def delete_test_file(bucket_id, file_id, name, retain_until):
        # Test cleanup only: the real service cannot expire compliance retention, so
        # advance the simulator's virtual clock past the file's retain-until time.
        advance = retain_until - virtual_now_ms() + 1000
        if advance > 0:
            simulator.advance_clock(advance)
            clock_offset_ms[0] += advance
        api.get_bucket_by_id(bucket_id).delete_file_version(file_id, name)

    class Handler(BaseHTTPRequestHandler):
        def log_message(self, *args):
            pass

        def do_POST(self):
            try:
                request = json.loads(
                    self.rfile.read(int(self.headers["Content-Length"]))
                )
                action = request["action"]
                if action == "provider":
                    if request["name"] != "bucket":
                        raise ValueError("unexpected provider operation")
                    result = json.loads(
                        Bucket(tool).run(
                            SimpleNamespace(OP=request["operation"]),
                            json.dumps(request["input"]),
                        )
                    )
                elif action == "create_file":
                    bucket = api.get_bucket_by_id(request["bucket_id"])
                    # A new retain-until must be after the simulator's virtual now.
                    retain_until = virtual_now_ms() + 3600000
                    version = bucket.upload_bytes(
                        b"retention test",
                        "retention-test.txt",
                        file_retention=FileRetentionSetting(
                            RetentionMode(request["mode"]), retain_until
                        ),
                    )
                    files[version.id_] = (bucket.id_, version.file_name, retain_until)
                    result = {
                        "file_id": version.id_,
                        "retention": api.get_file_info(
                            version.id_
                        ).file_retention.as_dict(),
                    }
                elif action == "get_file":
                    result = api.get_file_info(
                        request["file_id"]
                    ).file_retention.as_dict()
                elif action == "bucket_retention":
                    result = api.get_bucket_by_id(
                        request["bucket_id"]
                    ).default_retention.as_dict()
                elif action == "delete_file":
                    bucket_id, name, retain_until = files[request["file_id"]]
                    delete_test_file(bucket_id, request["file_id"], name, retain_until)
                    del files[request["file_id"]]
                    result = {"deleted": True}
                elif action == "bucket_exists":
                    from b2sdk.v3.exception import BucketIdNotFound

                    try:
                        api.get_bucket_by_id(request["bucket_id"])
                        result = {"exists": True}
                    except BucketIdNotFound:
                        result = {"exists": False}
                elif action == "cleanup_bucket":
                    from b2sdk.v3.exception import BucketIdNotFound

                    bucket_id = request["bucket_id"]
                    try:
                        bucket = api.get_bucket_by_id(bucket_id)
                        for file_id, (owner, name, retain_until) in list(files.items()):
                            if owner == bucket_id:
                                delete_test_file(bucket_id, file_id, name, retain_until)
                                del files[file_id]
                        api.delete_bucket(bucket)
                    except BucketIdNotFound:
                        pass
                    result = {"cleaned": True}
                else:
                    raise ValueError("unknown simulator action")
                output = json.dumps({"result": result}).encode()
                self.send_response(200)
            except Exception as exc:
                output = json.dumps({"error": str(exc)}).encode()
                self.send_response(500)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(output)))
            self.end_headers()
            self.wfile.write(output)

    server = HTTPServer(("127.0.0.1", 0), Handler)
    print(server.server_port, flush=True)
    try:
        server.serve_forever()
    finally:
        simulator.stop()


def invoke(url, data):
    req = urllib.request.Request(
        url,
        data=json.dumps(data).encode(),
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as response:
            payload = json.load(response)
    except urllib.error.HTTPError as exc:
        payload = json.load(exc)
        raise RuntimeError(payload.get("error", "simulator request failed")) from None
    return payload["result"]


if __name__ == "__main__":
    if len(sys.argv) == 2 and sys.argv[1] == "--serve":
        serve()
    else:
        try:
            result = invoke(
                os.environ["B2_RETENTION_SIMULATOR_URL"],
                {
                    "action": "provider",
                    "name": sys.argv[1],
                    "operation": sys.argv[2],
                    "input": json.load(sys.stdin),
                },
            )
            json.dump(result, sys.stdout)
        except Exception as exc:
            print(str(exc), file=sys.stderr)
            sys.exit(1)
