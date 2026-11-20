######################################################################
#
# File: python-bindings/tests/test_provider_tool.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

# Offline tests for b2_terraform/provider_tool.py (see conftest.py for how to run them and what the
# `known_defect` marker means).

import io
import json

import pytest
from b2sdk.v3.exception import BucketIdNotFound, FileNotPresent, ServiceError, Unauthorized

from b2_terraform import provider_tool
from b2_terraform.provider_tool import B2Provider, Bucket, BucketFiles, BucketFileVersion


def make_bucket(tool, name, file_lock=None):
    return Bucket(tool).resource_create(
        bucket_name=name,
        bucket_type='allPrivate',
        bucket_info={},
        cors_rules=[],
        file_lock_configuration=file_lock or [],
        default_server_side_encryption=[],
        lifecycle_rules=[],
    )


def upload(tool, bucket_id, file_name, data=b'hello'):
    bucket = tool.api.get_bucket_by_id(bucket_id)
    return bucket.upload_bytes(data, file_name)


def test_C11_default_retention_mode_none_creates_bucket(tool):
    """C11: default_retention mode 'none' (allowed by the Go schema) must not crash the python layer."""
    file_lock = [
        {
            'is_file_lock_enabled': True,
            'default_retention': [{'mode': 'none', 'period': []}],
        }
    ]

    result = make_bucket(tool, 'bucket-c11', file_lock)

    assert result['bucketId'], 'bucket was not created'


@pytest.mark.known_defect(raises=BucketIdNotFound, match='not found')
def test_C13_delete_of_already_deleted_bucket_succeeds(tool):
    """C13: destroy of a bucket that is already gone must be a no-op, not an error."""
    bucket = make_bucket(tool, 'bucket-c13')
    tool.api.delete_bucket(tool.api.get_bucket_by_id(bucket['bucketId']))

    Bucket(tool).resource_delete(bucket_id=bucket['bucketId'])  # must not raise


def test_C4_read_of_file_version_deleted_out_of_band_does_not_fail(tool, monkeypatch):
    """C4: refresh must tolerate a file version deleted outside Terraform (return an empty result).

    The simulator raises a bare KeyError for an unknown file id; real B2 answers 404 file_not_present, which
    b2sdk raises as FileNotPresent. The wrapper below makes the simulator behave like the service.
    """
    raw_api = tool.api.session.raw_api
    original = raw_api.get_file_info_by_id

    def like_the_service(api_url, account_auth_token, file_id):
        try:
            return original(api_url, account_auth_token, file_id)
        except KeyError:
            raise FileNotPresent(file_id_or_name=file_id) from None

    monkeypatch.setattr(raw_api, 'get_file_info_by_id', like_the_service)
    bucket = make_bucket(tool, 'bucket-c4')
    version = upload(tool, bucket['bucketId'], 'f.txt')
    tool.api.delete_file_version(version.id_, 'f.txt')

    result = BucketFileVersion(tool).resource_read(file_id=version.id_)

    assert not result, f'expected an empty result for a missing file version, got {result!r}'


@pytest.mark.parametrize(
    'error', [Unauthorized('access denied', 'unauthorized'), ServiceError('server failure')]
)
def test_file_version_read_propagates_auth_and_server_errors(tool, monkeypatch, error):
    def failing_read(file_id):
        raise error

    monkeypatch.setattr(tool.api, 'get_file_info', failing_read)
    with pytest.raises(type(error)) as caught:
        BucketFileVersion(tool).resource_read(file_id='existing-file')
    assert caught.value is error


def test_file_version_read_keeps_existing_attributes(tool):
    bucket = make_bucket(tool, 'bucket-existing-version')
    version = upload(tool, bucket['bucketId'], 'file.txt', b'data')

    result = BucketFileVersion(tool).resource_read(file_id=version.id_)

    assert result['fileId'] == version.id_
    assert result['fileName'] == 'file.txt'
    assert result['size'] == 4


def test_C39_non_recursive_listing_does_not_return_nested_files(tool):
    """C39: a non-recursive listing must not return files from sub-folders."""
    bucket = make_bucket(tool, 'bucket-c39')
    for name in ('top.txt', 'sub/a.txt', 'sub/b.txt', 'other/c.txt'):
        upload(tool, bucket['bucketId'], name)

    result = BucketFiles(tool).data_source_read(
        bucket_id=bucket['bucketId'], folder_name='', show_versions=False, recursive=False
    )

    names = sorted(v['fileName'] for v in result['fileVersions'])
    # retention first: a listing that returns nothing must fail here and not look like the known defect
    assert 'top.txt' in names, f'C39: the top-level file is missing from the listing: {names}'
    nested = [n for n in names if n.count('/') > 1 or (n.count('/') == 1 and not n.endswith('/'))]
    assert not nested, f'C39: non-recursive listing returned nested files: {names}'


@pytest.mark.known_defect(raises=AssertionError, match='C36: the surfaced error text')
def test_C36_failed_rollback_names_the_leftover_bucket(tool, monkeypatch):
    """C36: if updating retention fails and the rollback delete also fails, the error must name the
    bucket left behind so the user can clean it up."""
    file_lock = [
        {
            'is_file_lock_enabled': True,
            'default_retention': [
                {'mode': 'governance', 'period': [{'duration': 1, 'unit': 'days'}]}
            ],
        }
    ]

    def failing_update(self, **kwargs):
        raise ValueError('retention update failed')

    def failing_delete(bucket):
        raise RuntimeError('delete failed')

    monkeypatch.setattr(tool.api.BUCKET_CLASS, 'update', failing_update)
    monkeypatch.setattr(tool.api, 'delete_bucket', failing_delete)

    with pytest.raises(Exception) as excinfo:
        make_bucket(tool, 'bucket-c36', file_lock)

    surfaced = str(excinfo.value)
    assert 'retention update failed' in surfaced and 'bucket-c36' in surfaced, (
        'C36: the surfaced error text must carry both the original failure ("retention update failed") '
        f'and the name of the bucket left behind ("bucket-c36"), got: {surfaced!r}'
    )


@pytest.mark.known_defect(raises=AssertionError, match='C16: input was altered')
def test_C16_stdin_is_read_as_utf8_whatever_the_locale(monkeypatch):
    """C16: Go writes UTF-8 to stdin; on Windows the python child decoded it with the locale code
    page. Simulate a cp1252 locale and check that non-ASCII input arrives intact."""
    payload = {'bucket_id': 'b1', 'file_name': 'caf\u00e9-\u65e5\u672c.txt'}
    raw = json.dumps(payload, ensure_ascii=False).encode('utf-8')
    monkeypatch.setattr('sys.stdin', io.TextIOWrapper(io.BytesIO(raw + b'\n'), encoding='cp1252'))
    seen = []
    monkeypatch.setattr(B2Provider, 'run', lambda self, args, data_in: {})
    monkeypatch.setattr(
        provider_tool.AccountInfo, 'run', lambda self, args, data_in: seen.append(data_in) or '{}'
    )

    status = provider_tool.ProviderTool().run_command(['tool', 'account_info', 'data_source_read'])

    assert status == 0, f'run_command returned {status}'
    assert seen, 'the command did not receive any input'
    assert json.loads(seen[0]) == payload, f'C16: input was altered on the way in: {seen[0]!r}'
