######################################################################
#
# File: python-bindings/tests/test_non_recursive_listing.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import pytest

from b2_terraform.provider_tool import BucketFiles


def test_non_recursive_listing_returns_subfolder_entries(tool):
    bucket = tool.api.create_bucket('bucket-listing-regression', 'allPrivate')
    for name in ('top.txt', 'sub/a.txt', 'sub/b.txt', 'other/c.txt'):
        bucket.upload_bytes(b'content', name)

    result = BucketFiles(tool).data_source_read(
        bucket_id=bucket.id_, folder_name='', show_versions=False, recursive=False
    )

    entries = result['fileVersions']
    assert sorted(entry['fileName'] for entry in entries) == ['other/', 'sub/', 'top.txt']
    assert {entry['fileName']: entry['action'] for entry in entries} == {
        'other/': 'folder',
        'sub/': 'folder',
        'top.txt': 'upload',
    }
    folders = [entry for entry in entries if entry['action'] == 'folder']
    assert folders == [
        {'action': 'folder', 'bucketId': bucket.id_, 'fileName': 'other/'},
        {'action': 'folder', 'bucketId': bucket.id_, 'fileName': 'sub/'},
    ]


@pytest.mark.parametrize('show_versions', [False, True])
def test_non_recursive_listing_inside_folder_keeps_direct_files_and_prefixes(tool, show_versions):
    bucket = tool.api.create_bucket('bucket-listing-nested', 'allPrivate')
    direct = bucket.upload_bytes(b'direct', 'parent/direct.txt')
    bucket.upload_bytes(b'nested', 'parent/sub/nested.txt')
    bucket.upload_bytes(b'deeper', 'parent/sub/deeper/file.txt')
    bucket.upload_bytes(b'other', 'elsewhere/file.txt')

    result = BucketFiles(tool).data_source_read(
        bucket_id=bucket.id_, folder_name='parent/', show_versions=show_versions, recursive=False
    )

    entries = {entry['fileName']: entry for entry in result['fileVersions']}
    assert set(entries) == {'parent/direct.txt', 'parent/sub/'}
    assert entries['parent/direct.txt']['fileId'] == direct.id_
    assert entries['parent/sub/'] == {
        'action': 'folder',
        'bucketId': bucket.id_,
        'fileName': 'parent/sub/',
    }


@pytest.mark.parametrize('show_versions', [False, True])
def test_recursive_listing_preserves_nested_file_versions(tool, show_versions):
    bucket = tool.api.create_bucket('bucket-listing-recursive', 'allPrivate')
    first = bucket.upload_bytes(b'first', 'sub/file.txt')
    latest = bucket.upload_bytes(b'latest', 'sub/file.txt')
    other = bucket.upload_bytes(b'other', 'other/file.txt')

    result = BucketFiles(tool).data_source_read(
        bucket_id=bucket.id_, folder_name='', show_versions=show_versions, recursive=True
    )

    entries = result['fileVersions']
    expected_ids = {latest.id_, other.id_}
    if show_versions:
        expected_ids.add(first.id_)
    assert {entry['fileId'] for entry in entries} == expected_ids
    assert all(entry['action'] == 'upload' for entry in entries)
