######################################################################
#
# File: python-bindings/tests/test_default_retention_none.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import pytest

from b2_terraform.provider_tool import Bucket


@pytest.mark.parametrize('mode', ['governance', 'compliance'])
def test_default_retention_none_clears_existing_retention(tool, mode):
    bucket = Bucket(tool)
    created = bucket.resource_create(
        bucket_name='retention-none-regression',
        bucket_type='allPrivate',
        bucket_info={},
        cors_rules=[],
        file_lock_configuration=[
            {
                'is_file_lock_enabled': True,
                'default_retention': [
                    {
                        'mode': mode,
                        'period': [{'duration': 7, 'unit': 'days'}],
                    }
                ],
            }
        ],
        default_server_side_encryption=[],
        lifecycle_rules=[],
    )
    bucket_id = created['bucketId']
    before = tool.api.get_bucket_by_id(bucket_id).default_retention.as_dict()
    assert before == {'mode': mode, 'period': {'duration': 7, 'unit': 'days'}}

    bucket.resource_update(
        bucket_id=bucket_id,
        account_id=tool.api.account_info.get_account_id(),
        bucket_type='allPrivate',
        bucket_info={},
        cors_rules=[],
        file_lock_configuration=[
            {
                'is_file_lock_enabled': True,
                'default_retention': [{'mode': 'none', 'period': []}],
            }
        ],
        default_server_side_encryption=[],
        lifecycle_rules=[],
    )

    after = tool.api.get_bucket_by_id(bucket_id).default_retention.as_dict()
    assert after == {'mode': None}, f'default retention was not cleared: {after!r}'
