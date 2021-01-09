######################################################################
#
# File: python-bindings/tf_provider_b2/__main__.py
#
# Copyright 2021 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import json
import sys

from class_registry import ClassRegistry

from b2sdk.v1 import (
    # AuthInfoCache,
    B2Api,
    InMemoryAccountInfo,
    InMemoryCache,
    # SqliteAccountInfo,
)
from b2_terraform.arg_parser import ArgumentParser


def mixed_case_to_underscores(s):
    return s[0].lower() + ''.join(
        c if c.islower() or c.isdigit() else '_' + c.lower() for c in s[1:]
    )


class Command:
    # The registry for the subcommands, should be reinitialized  in subclass
    subcommands_registry = None

    def __init__(self, provider_tool):
        self.provider_tool = provider_tool
        self.api = provider_tool.api

    @classmethod
    def name(cls):
        return mixed_case_to_underscores(cls.__name__)

    @classmethod
    def register_subcommand(cls, command_class):
        assert cls.subcommands_registry is not None, 'Initialize the registry class'
        name = command_class.name()
        decorator = cls.subcommands_registry.register(key=name)(command_class)
        return decorator

    @classmethod
    def get_parser(cls, subparsers=None, parents=None):
        if parents is None:
            parents = []

        name = cls.name()
        if subparsers is None:
            parser = ArgumentParser(prog=name, parents=parents)
        else:
            parser = subparsers.add_parser(name, parents=parents)

        if cls.subcommands_registry:
            if not parents:
                common_parser = ArgumentParser(add_help=False)
                common_parser.add_argument('TYPE')
                common_parser.add_argument('CRUD')
                parents = [common_parser]

            subparsers = parser.add_subparsers(prog=parser.prog, title='usages', dest='command')
            subparsers.required = True
            for subcommand in cls.subcommands_registry.values():
                subcommand.get_parser(subparsers=subparsers, parents=parents)

        return parser

    def run(self, args, data_in):
        handler = getattr(self, f'{args.TYPE}_{args.CRUD}')
        data_out = json.dumps(handler(**json.loads(data_in)))
        return data_out


class B2Provider(Command):
    subcommands_registry = ClassRegistry()

    def run(self, args, data_in):
        self.provider_authorize_account(**json.loads(data_in))
        return {}

    def provider_authorize_account(self, _application_key_id, _application_key, **kwargs):
        if not _application_key_id or not _application_key:
            raise RuntimeError('B2 Application Key and Application Key ID must be provided')

        self.api.authorize_account('production', _application_key_id, _application_key)


@B2Provider.register_subcommand
class ApplicationKey(Command):
    def data_source_read(self, key_name, **kwargs):
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

    def resource_create(
        self, key_name, capabilities, bucket=None, duration=None, name_prefix=None, **kwargs
    ):
        if bucket is not None:
            bucket = self.api.get_bucket_by_name(bucket).id_

        response = self.api.create_key(
            capabilities=capabilities,
            key_name=key_name,
            valid_duration_seconds=duration,
            bucket_id=bucket,
            name_prefix=name_prefix,
        )

        try:
            return {
                'application_key_id': response['applicationKeyId'],
                'application_key': response['applicationKey'],
            }
        except KeyError:
            raise RuntimeError(f'Could not create Application Key for "{key_name}"')

    def resource_read(self, application_key_id, **kwargs):
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

    def resource_update(self, **kwargs):
        raise NotImplementedError(
            'Update is not available for Application Keys, every change requires recreation.'
        )

    def resource_delete(self, application_key_id, **kwargs):
        self.api.delete_key(application_key_id=application_key_id)

        return {}


class ProviderTool:
    def __init__(self, b2_api):
        self.api = b2_api

    def run_command(self, argv):
        try:
            data_in = input().strip()
            b2_provider = B2Provider(self)
            args = b2_provider.get_parser().parse_args(argv[1:])
            b2_provider.run(args, data_in)
            command_class = b2_provider.subcommands_registry.get_class(args.command)
            command = command_class(self)
            data_out = command.run(args, data_in)
            print(data_out, end='')
        except Exception as exc:
            print(exc, file=sys.stderr)
            return 1

        return 0


def main():
    # info = SqliteAccountInfo()
    # cache = AuthInfoCache(info)
    info = InMemoryAccountInfo()
    cache = InMemoryCache()
    # TODO: Append Terraform version to the User-Agent
    b2_api = B2Api(info, cache=cache)
    provider_tool = ProviderTool(b2_api=b2_api)
    return provider_tool.run_command(sys.argv)


if __name__ == '__main__':
    sys.exit(main())
