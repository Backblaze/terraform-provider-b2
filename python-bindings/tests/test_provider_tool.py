######################################################################
#
# File: python-bindings/tests/test_provider_tool.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

"""
Offline tests of the commands that run behind the Go provider. They need no B2 account.

Run them from the python-bindings directory with: python -m pytest tests
"""

from types import SimpleNamespace
from unittest import mock

import pytest
from b2sdk.v3.exception import BadRequest, BucketIdNotFound, FileNotPresent

from b2_terraform.provider_tool import Bucket, BucketFileVersion, BucketNotificationRules


def make_command(command_class):
    api = mock.MagicMock()
    return command_class(SimpleNamespace(api=api)), api


def test_bucket_delete_of_a_missing_bucket_is_a_no_op():
    command, api = make_command(Bucket)
    api.get_bucket_by_id.side_effect = BucketIdNotFound('bucket-id')

    assert command.resource_delete(bucket_id='bucket-id') is None
    api.delete_bucket.assert_not_called()


def test_bucket_delete_tolerates_a_bucket_removed_before_the_delete_call():
    command, api = make_command(Bucket)
    api.delete_bucket.side_effect = BadRequest('gone', 'bad_bucket_id')

    assert command.resource_delete(bucket_id='bucket-id') is None


def test_bucket_delete_reports_other_errors():
    command, api = make_command(Bucket)
    api.delete_bucket.side_effect = BadRequest('no way', 'some_other_code')

    with pytest.raises(BadRequest):
        command.resource_delete(bucket_id='bucket-id')


def test_notification_rules_delete_of_a_missing_bucket_is_a_no_op():
    # The behaviour the bucket resource is expected to share
    command, api = make_command(BucketNotificationRules)
    api.get_bucket_by_id.side_effect = BucketIdNotFound('bucket-id')

    assert command.resource_delete(bucket_id='bucket-id') is None


def test_file_version_read_of_a_missing_file_returns_none():
    command, api = make_command(BucketFileVersion)
    api.get_file_info.side_effect = FileNotPresent(file_id_or_name='file-id')

    assert command.resource_read(file_id='file-id') is None


@pytest.mark.skip(
    reason='Owner decision needed: it is unverified that B2 raises FileNotPresent when deleting a '
    'file version that is already gone, and Terraform does not call delete after a read drops the '
    'resource. Decide the intended behaviour, then enable this test.'
)
def test_file_version_delete_of_a_missing_file_is_a_no_op():
    command, api = make_command(BucketFileVersion)
    api.delete_file_version.side_effect = FileNotPresent(file_id_or_name='file-id')

    assert command.resource_delete(file_id='file-id', file_name='name') is None
