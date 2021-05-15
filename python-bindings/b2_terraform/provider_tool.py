######################################################################
#
# File: python-bindings/b2_terraform/provider_tool.py
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
from humps import camelize, decamelize
from b2sdk.v1 import EncryptionAlgorithm, EncryptionMode, EncryptionSetting

from b2_terraform.api_wrapper import B2ApiWrapper
from b2_terraform.arg_parser import ArgumentParser
from b2_terraform.json_encoder import B2ProviderJsonEncoder


def change_keys(obj, converter):
    return {converter(k).replace('__', '_'): v for k, v in obj.items()}


def apply_or_none(func, value):
    return None if value is None else func(value)


class Command:
    # The registry for the subcommands, should be reinitialized  in subclass
    subcommands_registry = None

    def __init__(self, provider_tool):
        self.provider_tool = provider_tool
        self.api = provider_tool.api

    @classmethod
    def name(cls):
        return decamelize(cls.__name__)

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
        result['_ua'] = self.api.user_agent
        data_out = json.dumps(
            change_keys(result, converter=decamelize),
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
        return self._postprocess(**bucket.as_dict())

    def resource_create(
        self,
        *,
        bucket_name,
        bucket_type,
        bucket_info,
        cors_rules,
        file_lock_configuration,
        default_server_side_encryption,
        lifecycle_rules,
        **kwargs,
    ):
        bucket = self.api.create_bucket(
            **self._preprocess(
                name=bucket_name,
                bucket_type=bucket_type,
                bucket_info=bucket_info,
                cors_rules=cors_rules,
                # file_lock_configuration=file_lock_configuration,
                default_server_side_encryption=default_server_side_encryption,
                lifecycle_rules=lifecycle_rules,
            )
        )
        return self._postprocess(**bucket.as_dict())

    def resource_read(self, *, bucket_id, **kwargs):
        bucket = self.api.get_bucket_by_id(bucket_id)
        return self._postprocess(**bucket.as_dict())

    def resource_update(
        self,
        bucket_id,
        account_id,
        bucket_type,
        bucket_info,
        cors_rules,
        file_lock_configuration,
        default_server_side_encryption,
        lifecycle_rules,
        **kwargs,
    ):
        self.api.session.update_bucket(
            **self._preprocess(
                account_id=account_id,
                bucket_id=bucket_id,
                bucket_type=bucket_type,
                bucket_info=bucket_info,
                cors_rules=cors_rules,
                file_lock_configuration=file_lock_configuration,
                default_server_side_encryption=default_server_side_encryption,
                lifecycle_rules=lifecycle_rules,
            )
        )
        bucket = self.api.get_bucket_by_id(bucket_id)
        return self._postprocess(**bucket.as_dict())

    def resource_delete(self, *, bucket_id, **kwargs):
        bucket = self.api.get_bucket_by_id(bucket_id)
        self.api.delete_bucket(bucket)

        return {}

    def _preprocess(self, **kwargs):
        cors_rules = kwargs.pop('cors_rules')
        if cors_rules:
            for index, item in enumerate(cors_rules):
                cors_rules[index] = change_keys(item, converter=camelize)

        # TODO: filelock

        default_server_side_encryption = kwargs.pop('default_server_side_encryption')
        if default_server_side_encryption:
            mode = apply_or_none(EncryptionMode, default_server_side_encryption[0]['mode'] or None)
            if mode:
                algorithm = apply_or_none(
                    EncryptionAlgorithm, default_server_side_encryption[0]['algorithm'] or None
                )
                default_server_side_encryption = EncryptionSetting(mode=mode, algorithm=algorithm)
            else:
                default_server_side_encryption = None
        else:
            default_server_side_encryption = None

        lifecycle_rules = kwargs.pop('lifecycle_rules')
        if lifecycle_rules:
            for index, item in enumerate(lifecycle_rules):
                days_from_hiding_to_deleting = item.get('days_from_hiding_to_deleting')
                if days_from_hiding_to_deleting == 0:
                    item['days_from_hiding_to_deleting'] = None
                days_from_uploading_to_hiding = item.get('days_from_uploading_to_hiding')
                if days_from_uploading_to_hiding == 0:
                    item['days_from_uploading_to_hiding'] = None
                lifecycle_rules[index] = change_keys(item, converter=camelize)

        return {
            'cors_rules': cors_rules,
            'default_server_side_encryption': default_server_side_encryption,
            'lifecycle_rules': lifecycle_rules,
            **kwargs,
        }

    def _postprocess(self, **kwargs):
        cors_rules = kwargs.pop('corsRules')
        if cors_rules:
            for index, item in enumerate(cors_rules):
                cors_rules[index] = change_keys(item, converter=decamelize)
        else:
            cors_rules = []

        # TODO: filelock

        default_server_side_encryption = kwargs.pop('defaultServerSideEncryption')
        if default_server_side_encryption:
            default_server_side_encryption = [
                change_keys(default_server_side_encryption, converter=decamelize)
            ]
        else:
            default_server_side_encryption = []

        lifecycle_rules = kwargs.pop('lifecycleRules')
        if lifecycle_rules:
            for index, item in enumerate(lifecycle_rules):
                lifecycle_rules[index] = change_keys(item, converter=decamelize)
        else:
            lifecycle_rules = []

        return {
            'corsRules': cors_rules,
            'defaultServerSideEncryption': default_server_side_encryption,
            'lifecycleRules': lifecycle_rules,
            **kwargs,
        }


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
        server_side_encryption,
        **kwargs,
    ):
        bucket = self.api.get_bucket_by_id(bucket_id)
        file_info = bucket.upload_local_file(
            **self._preprocess(
                local_file=source,
                file_name=file_name,
                content_type=content_type,
                file_infos=file_info,
                server_side_encryption=server_side_encryption,
            ),
        )
        return self._postprocess(bucketId=bucket_id, source=source, **file_info.as_dict())

    def resource_read(self, *, file_id, **kwargs):
        return self._postprocess(**self.api.get_file_info(file_id))

    def resource_delete(self, *, file_id, file_name, **kwargs):
        self.api.delete_file_version(file_id, file_name)

        return {}

    def _preprocess(self, **kwargs):
        content_type = kwargs.pop('content_type') or None

        server_side_encryption = kwargs.pop('server_side_encryption')
        if server_side_encryption:
            mode = apply_or_none(EncryptionMode, server_side_encryption[0]['mode'] or None)
            if mode:
                algorithm = apply_or_none(
                    EncryptionAlgorithm, server_side_encryption[0]['algorithm'] or None
                )
                server_side_encryption = EncryptionSetting(mode=mode, algorithm=algorithm)
            else:
                server_side_encryption = None
        else:
            server_side_encryption = None

        return {
            'content_type': content_type,
            'encryption': server_side_encryption,
            **kwargs,
        }

    def _postprocess(self, **kwargs):
        server_side_encryption = kwargs.pop('serverSideEncryption', None)
        if server_side_encryption:
            server_side_encryption = [change_keys(server_side_encryption, converter=decamelize)]
        else:
            server_side_encryption = []

        return {
            'serverSideEncryption': server_side_encryption,
            **kwargs,
        }


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
        return self._postprocess(
            bucketId=bucket_id, fileName=file_name, showVersions=show_versions, generator=generator
        )

    def _postprocess(self, *, generator, **kwargs):
        def postprocess_file_version(**kwargs):
            server_side_encryption = kwargs.pop('serverSideEncryption', None)
            if server_side_encryption:
                server_side_encryption = [change_keys(server_side_encryption, converter=decamelize)]
            else:
                server_side_encryption = []

            return {
                'serverSideEncryption': server_side_encryption,
                **kwargs,
            }

        file_name = kwargs['fileName']
        file_versions = [
            change_keys(
                postprocess_file_version(**file_version_info.as_dict()), converter=decamelize
            )
            for file_version_info, _ in generator
            if file_version_info.file_name == file_name
        ]

        return {
            'fileVersions': file_versions,
            **kwargs,
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
        return self._postprocess(
            bucketId=bucket_id,
            folderName=folder_name,
            showVersions=show_versions,
            recursive=recursive,
            generator=generator,
        )

    def _postprocess(self, *, generator, **kwargs):
        def postprocess_file_version(**kwargs):
            server_side_encryption = kwargs.pop('serverSideEncryption', None)
            if server_side_encryption:
                server_side_encryption = [change_keys(server_side_encryption, converter=decamelize)]
            else:
                server_side_encryption = []

            return {
                'serverSideEncryption': server_side_encryption,
                **kwargs,
            }

        file_versions = [
            change_keys(
                postprocess_file_version(**file_version_info.as_dict()), converter=decamelize
            )
            for file_version_info, _ in generator
        ]

        return {
            'fileVersions': file_versions,
            **kwargs,
        }


@B2Provider.register_subcommand
class AccountInfo(Command):
    def data_source_read(self, **kwargs):
        account_info = self.api.account_info
        return {
            'accountId': account_info.get_account_id(),
            'allowed': [change_keys(account_info.get_allowed(), converter=decamelize)],
            'accountAuthToken': account_info.get_account_auth_token(),
            'apiUrl': account_info.get_api_url(),
            'downloadUrl': account_info.get_download_url(),
            's3_ApiUrl': account_info.get_s3_api_url(),
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
