######################################################################
#
# File: python-bindings/b2_terraform/terraform_structures.py
#
# Copyright 2021 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

"""
Store information about which data keys returned by B2 SDK should be passed on to terraform provider.
Prevents error in terraform provider when API is extended with new fields.
"""

BUCKET_SERVER_SIDE_ENCRYPTION = {
    "mode": True,
    "algorithm": True,
}

FILE_VERSION_SERVER_SIDE_ENCRYPTION = {
    "mode": True,
    "algorithm": True,
    # 'key' is not here as API does not return it
}

FILE_VERSION_KEYS = {
    "action": True,
    "content_md5": True,
    "content_sha1": True,
    "content_type": True,
    "bucket_id": True,
    "file_id": True,
    "file_info": True,
    "file_name": True,
    "size": True,
    "source": True,
    "server_side_encryption": FILE_VERSION_SERVER_SIDE_ENCRYPTION,
    "upload_timestamp": True,
}

FILE_KEYS = {
    "bucket_id": True,
    "file_name": True,
    "show_versions": True,
    "file_versions": FILE_VERSION_KEYS,
}

FILE_SIGNED_URL_KEYS = {
    "bucket_id": True,
    "file_name": True,
    "duration": True,
    "signed_url": True,
}

FILES_KEYS = {
    "bucket_id": True,
    "folder_name": True,
    "show_versions": True,
    "recursive": True,
    "file_versions": FILE_VERSION_KEYS,
}

BUCKET_KEYS = {
    "bucket_id": True,
    "bucket_name": True,
    "bucket_type": True,
    "bucket_info": True,
    "cors_rules": {
        "cors_rule_name": True,
        "allowed_origins": True,
        "allowed_operations": True,
        "max_age_seconds": True,
        "allowed_headers": True,
        "expose_headers": True,
    },
    "file_lock_configuration": {
        "is_file_lock_enabled": True,
        "default_retention": {
            "mode": True,
            "period": {
                "duration": True,
                "unit": True,
            },
        },
    },
    "default_server_side_encryption": BUCKET_SERVER_SIDE_ENCRYPTION,
    "lifecycle_rules": {
        "file_name_prefix": True,
        "days_from_hiding_to_deleting": True,
        "days_from_uploading_to_hiding": True,
    },
    "account_id": True,
    "options": True,
    "revision": True,
}
