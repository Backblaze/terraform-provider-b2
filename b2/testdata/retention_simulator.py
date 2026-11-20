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
"""Persistent b2sdk simulator for the Terraform retention acceptance test.

RawSimulator has no default-retention inheritance on upload. The test therefore
creates a file with explicit retention and compares its server-side mode and
timestamp before and after changing the bucket default. After the assertions,
cleanup expires the simulator's file retention to allow deletion.
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
    B2HttpApiConfig,
    DummyCache,
    FileRetentionSetting,
    InMemoryAccountInfo,
    RawSimulator,
    RetentionMode,
)

sys.path.insert(
    0, os.path.join(os.path.dirname(__file__), "..", "..", "python-bindings")
)
from b2_terraform.provider_tool import Bucket, ProviderTool


def serve():
    tool = ProviderTool()
    api = B2Api(
        account_info=InMemoryAccountInfo(),
        cache=DummyCache(),
        api_config=B2HttpApiConfig(_raw_api_class=RawSimulator),
    )
    key_id, key = api.session.raw_api.create_account()
    api.authorize_account(key_id, key)
    tool.__dict__["api"] = api
    files = {}

    def delete_test_file(bucket_id, file_id, name):
        # Test cleanup only: the real service cannot expire compliance retention.
        stored = api.session.raw_api.bucket_id_to_bucket[bucket_id].file_id_to_file[
            file_id
        ]
        # RawSimulator's deletion check compares this timestamp with seconds.
        stored.file_retention.retain_until = int(time.time()) - 1
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
                    retain_until = int(time.time() * 1000) + 3600000
                    version = bucket.upload_bytes(
                        b"retention test",
                        "retention-test.txt",
                        file_retention=FileRetentionSetting(
                            RetentionMode(request["mode"]), retain_until
                        ),
                    )
                    files[version.id_] = (bucket.id_, version.file_name)
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
                    bucket_id, name = files[request["file_id"]]
                    delete_test_file(bucket_id, request["file_id"], name)
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
                        for file_id, (owner, name) in list(files.items()):
                            if owner == bucket_id:
                                delete_test_file(bucket_id, file_id, name)
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
    server.serve_forever()


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
