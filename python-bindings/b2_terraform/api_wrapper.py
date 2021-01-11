######################################################################
#
# File: python-bindings/tf_provider_b2/api.py
#
# Copyright 2021 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

from b2sdk.v1 import B2Api, InMemoryAccountInfo


class B2ApiWrapper(B2Api):
    def __init__(self):
        account_info = InMemoryAccountInfo()
        # TODO: Append Terraform version to the User-Agent
        super().__init__(account_info=account_info)

    def get_bucket_by_id(self, bucket_id):
        # INFO: in Terraform we must ask the API.

        for bucket in self.list_buckets(bucket_id=bucket_id):
            assert bucket.id_ == bucket_id
            return bucket

        # There is no such bucket.
        raise NonExistentBucket(bucket_id)

    def list_buckets(self, bucket_name=None, bucket_id=None):
        # INFO: added bucket_id argument. Remove it when SDK support it.

        # Give a useful warning if the current application key does not
        # allow access to the named bucket.
        self.check_bucket_restrictions(bucket_name)

        account_id = self.account_info.get_account_id()
        self.check_bucket_restrictions(bucket_name)

        response = self.session.list_buckets(
            account_id, bucket_name=bucket_name, bucket_id=bucket_id
        )
        buckets = self.BUCKET_FACTORY_CLASS.from_api_response(self, response)

        if bucket_name is not None:
            # If a bucket_name is specified we don't clear the cache because the other buckets could still
            # be valid. So we save the one bucket returned from the list_buckets call.
            for bucket in buckets:
                self.cache.save_bucket(bucket)
        else:
            # Otherwise we want to clear the cache and save the buckets returned from list_buckets
            # since we just got a new list of all the buckets for this account.
            self.cache.set_bucket_name_cache(buckets)

        return buckets
