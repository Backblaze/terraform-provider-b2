######################################################################
#
# File: python-bindings/tests/test_bucket_files_limit.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import pytest

from b2_terraform.provider_tool import BucketFiles


def test_bucket_files_rejects_results_over_limit(tool):
    bucket = tool.api.create_bucket('listing-limit-regression', 'allPrivate')
    for name in ('a.txt', 'b.txt', 'c.txt'):
        bucket.upload_bytes(b'content', name)

    with pytest.raises(ValueError, match='max_results'):
        BucketFiles(tool).data_source_read(
            bucket_id=bucket.id_, folder_name='', show_versions=False, recursive=True, max_results=2
        )


@pytest.mark.parametrize('count', [0, 1, 2])
def test_listing_at_or_below_limit_returns_all_entries(tool, count):
    bucket = tool.api.create_bucket('listing-limit-boundaries', 'allPrivate')
    for index in range(count):
        bucket.upload_bytes(b'content', f'{index}.txt')
    result = BucketFiles(tool).data_source_read(
        bucket_id=bucket.id_, folder_name='', show_versions=False, recursive=True, max_results=2
    )
    assert len(result['fileVersions']) == count
    assert result['maxResults'] == 2


@pytest.mark.parametrize('value', [0, -1, 100001, True, '2'])
def test_listing_rejects_invalid_limit(tool, value):
    with pytest.raises(ValueError, match='max_results'):
        BucketFiles(tool).data_source_read(
            bucket_id='unused',
            folder_name='',
            show_versions=False,
            recursive=False,
            max_results=value,
        )


def test_listing_consumes_at_most_one_entry_beyond_limit(tool, monkeypatch):
    class BucketStub:
        calls = 0
        page_size = None

        def ls(self, folder_name, *, latest_only, recursive, fetch_count):
            self.page_size = fetch_count
            for index in range(100):
                self.calls += 1
                yield {'fileName': str(index)}, None

    bucket = BucketStub()
    monkeypatch.setattr(tool.api, 'get_bucket_by_id', lambda _: bucket)
    with pytest.raises(ValueError, match='max_results'):
        BucketFiles(tool).data_source_read(
            bucket_id='unused', folder_name='', show_versions=False, recursive=True, max_results=3
        )
    assert bucket.calls == 4
    assert bucket.page_size == 4


@pytest.mark.parametrize('show_versions', [False, True])
def test_limit_counts_file_versions_and_folder_entries(tool, show_versions):
    bucket = tool.api.create_bucket('listing-limit-folders', 'allPrivate')
    bucket.upload_bytes(b'one', 'root.txt')
    bucket.upload_bytes(b'two', 'root.txt')
    bucket.upload_bytes(b'one', 'folder/a.txt')
    bucket.upload_bytes(b'two', 'folder/b.txt')
    command = BucketFiles(tool)
    with pytest.raises(ValueError, match='max_results'):
        command.data_source_read(
            bucket_id=bucket.id_,
            folder_name='',
            show_versions=show_versions,
            recursive=False,
            max_results=1,
        )
    expected_count = 3 if show_versions else 2
    result = command.data_source_read(
        bucket_id=bucket.id_,
        folder_name='',
        show_versions=show_versions,
        recursive=False,
        max_results=expected_count,
    )
    assert len(result['fileVersions']) == expected_count
    assert sum(entry['action'] == 'folder' for entry in result['fileVersions']) == 1
