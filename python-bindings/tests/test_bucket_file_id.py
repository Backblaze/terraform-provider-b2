######################################################################
#
# File: python-bindings/tests/test_bucket_file_id.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import json
from types import SimpleNamespace

import pytest

from b2_terraform.provider_tool import BucketFile, BucketFiles


@pytest.mark.parametrize('command', [BucketFile, BucketFiles])
def test_bucket_file_response_omits_input_digest(tool, command):
    bucket = tool.api.create_bucket('bucket-file-id-regression', 'allPrivate')
    bucket.upload_bytes(b'content', 'file.txt')
    request = {
        'bucket_id': bucket.id_,
        'file_name': 'file.txt',
        'show_versions': False,
        'folder_name': '',
        'recursive': False,
        'provider_application_key_id': 'example-id',
        'provider_application_key': 'example-key',
        'provider_endpoint': 'production',
    }
    result = json.loads(
        command(tool).run(SimpleNamespace(OP='data_source_read'), json.dumps(request))
    )
    assert result['bucketId'] == bucket.id_
    assert '_sha1' not in result
