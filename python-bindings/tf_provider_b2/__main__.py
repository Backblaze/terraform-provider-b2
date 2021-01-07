######################################################################
#
# File: python-bindings/tf_provider_b2/__main__.py
#
# Copyright 2020 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################


import argparse
import json
import sys

from b2sdk.v1 import (
    # AuthInfoCache,
    B2Api,
    InMemoryAccountInfo,
    InMemoryCache,
    # SqliteAccountInfo,
)


class ThrowingArgumentParser(argparse.ArgumentParser):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, add_help=False, **kwargs)

    def error(self, message):
        raise RuntimeError(message)


def parse_args():
    parser = ThrowingArgumentParser()
    parser.add_argument('TYPE')
    parser.add_argument('NAME')
    parser.add_argument('CRUD')

    return parser.parse_args()


class B2Dispatcher:
    def __init__(self):
        # info = SqliteAccountInfo()
        # cache = AuthInfoCache(info)
        info = InMemoryAccountInfo()
        cache = InMemoryCache()
        self.api = B2Api(info, cache=cache)

    def dispatch(self, type_, name, crud, data_in):
        handler = getattr(self, f'{type_}_{name}_{crud}')
        data_out = json.dumps(handler(**json.loads(data_in)))
        return data_out

    def provider_authorize_account(self, _application_key_id, _application_key):
        if not _application_key_id or not _application_key:
            raise RuntimeError('B2 Application Key and Application Key ID must be provided')

        self.api.authorize_account('production', _application_key_id, _application_key)

    def data_source_application_key_id_read(self, key_name, **kwargs):
        self.provider_authorize_account(**kwargs)

        next_id = None
        while True:
            response = self.api.list_keys(next_id)
            keys = response['keys']

            for key in keys:
                if key_name == key['keyName']:
                    return {
                        'application_key_id': key['applicationKeyId'],
                        'capabilities': key['capabilities'],
                    }

            next_id = response.get('nextApplicationKeyId')
            if next_id is None:
                break

        raise RuntimeError(f'Could not find Application Key ID for "{key_name}"')

    def resource_application_key_id_create(self, key_name, capabilities, bucket=None, duration=None, name_prefix=None, **kwargs):
        self.provider_authorize_account(**kwargs)

        if bucket is not None:
            bucket = self.api.get_bucket_by_name(bucket).id_

        response = self.api.create_key(
            capabilities=capabilities,
            key_name=key_name,
            valid_duration_seconds=duration,
            bucket_id=bucket,
            name_prefix=name_prefix
        )

        try:
            return {
                'application_key_id': response['applicationKeyId'],
                'application_key': response['applicationKey'],
            }
        except KeyError:
            raise RuntimeError(f'Could not create Application Key for "{key_name}"')

    def resource_application_key_id_read(self, application_key_id, **kwargs):
        self.provider_authorize_account(**kwargs)

        next_id = None
        while True:
            response = self.api.list_keys(next_id)
            keys = response['keys']

            for key in keys:
                if application_key_id == key['applicationKeyId']:
                    return {
                        'key_name': key['keyName'],
                        'capabilities': key['capabilities'],
                    }

            next_id = response.get('nextApplicationKeyId')
            if next_id is None:
                break

        raise RuntimeError(f'Could not find Application Key ID for ID "{application_key_id}"')

    def resource_application_key_id_update(self, **kwargs):
        raise NotImplementedError('Update is not available for API Keys, every change requires recreation.')

    def resource_application_key_id_delete(self, application_key_id, **kwargs):
        self.provider_authorize_account(**kwargs)

        self.api.delete_key(application_key_id=application_key_id)

        return {}


def main():
    try:
        args = parse_args()
        dispatcher = B2Dispatcher()

        data_in = input().strip()
        data_out = dispatcher.dispatch(args.TYPE, args.NAME, args.CRUD, data_in)
        print(data_out, end='')
    except Exception as exc:
        print(exc, file=sys.stderr)
        return 1

    return 0


if __name__ == '__main__':
    sys.exit(main())
