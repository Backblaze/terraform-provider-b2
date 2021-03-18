######################################################################
#
# File: python-bindings/b2_terraform/api_wrapper.py
#
# Copyright 2021 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import os

from b2sdk.v1 import B2Api, B2Http, B2RawApi, InMemoryAccountInfo
from b2sdk.v1.exception import NonExistentBucket


class B2ApiWrapper(B2Api):
    def __init__(self):
        account_info = InMemoryAccountInfo()
        raw_api = B2RawApi(B2Http(user_agent_append=os.environ.get('B2_USER_AGENT_APPEND')))

        super().__init__(account_info=account_info, raw_api=raw_api)

    @property
    def user_agent(self):
        # TODO: Remove it when SDK has it.
        return self.raw_api.b2_http.user_agent

    def get_bucket_by_id(self, bucket_id):
        # INFO: in Terraform we must ask the API.

        for bucket in self.list_buckets(bucket_id=bucket_id):
            assert bucket.id_ == bucket_id
            return bucket

        # There is no such bucket.
        raise NonExistentBucket(bucket_id)
