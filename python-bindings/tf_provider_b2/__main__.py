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
    def error(self, message):
        raise RuntimeError(message)


def parse_args():
    parser = ThrowingArgumentParser()
    parser.add_argument('TYPE')
    parser.add_argument('NAME')

    return parser.parse_args()


class B2Dispatcher:
    def __init__(self):
        # info = SqliteAccountInfo()
        # cache = AuthInfoCache(info)
        info = InMemoryAccountInfo()
        cache = InMemoryCache()
        self.api = B2Api(info, cache=cache)

    def dispatch(self, type_, name, data_in):
        handler = getattr(self, f'{type_}_{name}')
        data_out = json.dumps(handler(**json.loads(data_in)))
        return data_out

    def provider_authorize_account(self, _application_key_id, _application_key):
        if not _application_key_id or not _application_key:
            raise RuntimeError('B2 Application Key and Application Key ID must be provided')

        self.api.authorize_account('production', _application_key_id, _application_key)

    def data_source_application_key_id(self, key_name, **kwargs):
        self.provider_authorize_account(**kwargs)

        next_id = None

        while True:
            response = self.api.list_keys(next_id)
            keys = response['keys']

            for key in keys:
                if key_name == key['keyName']:
                    return {'application_key_id': key['applicationKeyId']}

            next_id = response.get('nextApplicationKeyId')
            if next_id is None:
                break

        raise RuntimeError(f'Could not find Application Key ID for "{key_name}"')


def main():
    try:
        args = parse_args()
        dispatcher = B2Dispatcher()

        data_in = input().strip()
        data_out = dispatcher.dispatch(args.TYPE, args.NAME, data_in)
        print(data_out, end='')
    except Exception as exc:
        print(exc, file=sys.stderr)
        return 1

    return 0


if __name__ == '__main__':
    sys.exit(main())
