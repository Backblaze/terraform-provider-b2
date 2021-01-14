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
import hashlib
import os
import sys

from class_registry import ClassRegistry

from b2_terraform.api_wrapper import B2ApiWrapper
from b2_terraform.arg_parser import ArgumentParser
from b2_terraform.json_encoder import B2ProviderJsonEncoder


def mixed_case_to_underscores(s):
    return s[0].lower() + ''.join(
        c if c.islower() or c.isdigit() else '_' + c.lower() for c in s[1:]
    )


def change_keys(obj, converter):
    return {converter(k): v for k, v in obj.items()}


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
                common_parser.add_argument('OP')
                parents = [common_parser]

            subparsers = parser.add_subparsers(prog=parser.prog, title='usages', dest='CMD')
            subparsers.required = True
            for subcommand in cls.subcommands_registry.values():
                subcommand.get_parser(subparsers=subparsers, parents=parents)

        return parser

    def run(self, args, data_in):
        handler = getattr(self, args.OP)
        result = handler(**json.loads(data_in))
        result['_sha1'] = hashlib.sha1(data_in.encode()).hexdigest()
        data_out = json.dumps(
            change_keys(result, converter=mixed_case_to_underscores),
            cls=B2ProviderJsonEncoder,
            sort_keys=True,
        )
        return data_out


class B2Provider(Command):
    subcommands_registry = ClassRegistry()

    def run(self, args, data_in):
        self.provider_authorize_account(**json.loads(data_in))
        return {}

    def provider_authorize_account(
        self, *, provider_application_key_id, provider_application_key, provider_endpoint, **kwargs
    ):
        if not provider_application_key_id or not provider_application_key:
            raise RuntimeError('B2 Application Key and Application Key ID must be provided')

        self.api.authorize_account(
            provider_endpoint, provider_application_key_id, provider_application_key
        )


@B2Provider.register_subcommand
class ApplicationKey(Command):
    def data_source_read(self, *, key_name, **kwargs):
        next_id = None
        while True:
            response = self.api.list_keys(next_id)
            keys = response['keys']

            for key in keys:
                if key_name == key['keyName']:
                    return key

            next_id = response.get('nextApplicationKeyId')
            if next_id is None:
                break

        raise RuntimeError(f'Could not find Application Key for "{key_name}"')

    def resource_create(self, *, key_name, capabilities, bucket_id, name_prefix, **kwargs):
        return self.api.create_key(
            key_name=key_name,
            capabilities=capabilities,
            bucket_id=bucket_id or None,
            name_prefix=name_prefix or None,
        )

    def resource_read(self, *, application_key_id, **kwargs):
        next_id = None
        while True:
            response = self.api.list_keys(next_id)
            keys = response['keys']

            for key in keys:
                if application_key_id == key['applicationKeyId']:
                    return key

            next_id = response.get('nextApplicationKeyId')
            if next_id is None:
                break

        raise RuntimeError(f'Could not find Application Key for ID "{application_key_id}"')

    def resource_delete(self, *, application_key_id, **kwargs):
        self.api.delete_key(application_key_id=application_key_id)

        return {}


@B2Provider.register_subcommand
class Bucket(Command):
    def data_source_read(self, *, bucket_name, **kwargs):
        bucket = self.api.get_bucket_by_name(bucket_name)
        return bucket.as_dict()

    def resource_create(
        self,
        *,
        bucket_name,
        bucket_type,
        bucket_info=None,
        cors_rules=None,
        lifecycle_rules=None,
        **kwargs,
    ):
        bucket = self.api.create_bucket(
            bucket_name,
            bucket_type,
            bucket_info=bucket_info,
            cors_rules=cors_rules,
            lifecycle_rules=lifecycle_rules,
        )
        return bucket.as_dict()

    def resource_read(self, *, bucket_id, **kwargs):
        bucket = self.api.get_bucket_by_id(bucket_id)
        return bucket.as_dict()

    def resource_update(
        self, bucket_id, account_id, bucket_type, bucket_info, cors_rules, lifecycle_rules, **kwargs
    ):
        self.api.session.update_bucket(
            account_id,
            bucket_id,
            bucket_type=bucket_type,
            bucket_info=bucket_info,
            cors_rules=cors_rules,
            lifecycle_rules=lifecycle_rules,
        )
        bucket = self.api.get_bucket_by_id(bucket_id)
        return bucket.as_dict()

    def resource_delete(self, *, bucket_id, **kwargs):
        bucket = self.api.get_bucket_by_id(bucket_id)
        self.api.delete_bucket(bucket)

        return {}


@B2Provider.register_subcommand
class BucketFileVersion(Command):
    def resource_create(
        self,
        *,
        bucket_id,
        file_name,
        source,
        content_type,
        file_info,
        **kwargs,
    ):
        bucket = self.api.get_bucket_by_id(bucket_id)
        file_info = bucket.upload_local_file(
            local_file=source,
            file_name=file_name,
            content_type=content_type or None,
            file_infos=file_info,
        )
        result = file_info.as_dict()
        result['bucketId'] = bucket_id
        result['source'] = source
        return result

    def resource_read(self, *, file_id, **kwargs):
        return self.api.get_file_info(file_id)

    def resource_delete(self, *, file_id, file_name, **kwargs):
        self.api.delete_file_version(file_id, file_name)

        return {}


@B2Provider.register_subcommand
class BucketFile(Command):
    def data_source_read(self, *, bucket_id, file_name, show_versions, **kwargs):
        bucket = self.api.get_bucket_by_id(bucket_id)
        folder_name = os.path.dirname(file_name)
        generator = bucket.ls(
            folder_name,
            show_versions=show_versions,
            recursive=False,
        )
        return {
            'bucketId': bucket_id,
            'fileName': file_name,
            'showVersions': show_versions,
            'fileVersions': [
                change_keys(file_version_info.as_dict(), converter=mixed_case_to_underscores)
                for file_version_info, _ in generator
                if file_version_info.file_name == file_name
            ],
        }


@B2Provider.register_subcommand
class BucketFiles(Command):
    def data_source_read(self, *, bucket_id, folder_name, show_versions, recursive, **kwargs):
        bucket = self.api.get_bucket_by_id(bucket_id)
        generator = bucket.ls(
            folder_name,
            show_versions=show_versions,
            recursive=recursive,
        )
        return {
            'bucketId': bucket_id,
            'folderName': folder_name,
            'showVersions': show_versions,
            'recursive': recursive,
            'fileVersions': [
                change_keys(file_version_info.as_dict(), converter=mixed_case_to_underscores)
                for file_version_info, _ in generator
            ],
        }


class ProviderTool:
    def __init__(self, b2_api):
        self.api = b2_api

    def run_command(self, argv):
        try:
            b2_provider = B2Provider(self)
            args = b2_provider.get_parser().parse_args(argv[1:])
            data_in = input().strip()
            b2_provider.run(args, data_in)
            command_class = b2_provider.subcommands_registry.get_class(args.CMD)
            command = command_class(self)
            data_out = command.run(args, data_in)
            print(data_out, end='')
        except Exception as exc:
            print(exc, file=sys.stderr)
            return 1

        return 0


def main():
    b2_api = B2ApiWrapper()
    provider_tool = ProviderTool(b2_api=b2_api)
    return provider_tool.run_command(sys.argv)


if __name__ == '__main__':
    sys.exit(main())
