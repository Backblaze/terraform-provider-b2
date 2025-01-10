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

API_KEY_KEYS = {
    "application_key": None,
    "application_key_id": None,
    "bucket_id": "",
    "capabilities": None,
    "name_prefix": "",
    "options": None,
    "key_name": None,
}

BUCKET_SERVER_SIDE_ENCRYPTION = {
    "mode": None,
    "algorithm": None,
}

FILE_VERSION_SERVER_SIDE_ENCRYPTION = {
    "mode": None,
    "algorithm": None,
    # 'key' is not here as API does not return it
}

FILE_VERSION_KEYS = {
    "action": None,
    "content_md5": None,
    "content_sha1": None,
    "content_type": None,
    "bucket_id": None,
    "file_id": None,
    "file_info": None,
    "file_name": None,
    "size": None,
    "source": None,
    "server_side_encryption": FILE_VERSION_SERVER_SIDE_ENCRYPTION,
    "upload_timestamp": None,
}

FILE_KEYS = {
    "bucket_id": None,
    "file_name": None,
    "show_versions": None,
    "file_versions": FILE_VERSION_KEYS,
}

FILE_SIGNED_URL_KEYS = {
    "bucket_id": None,
    "file_name": None,
    "duration": None,
    "signed_url": None,
}

FILES_KEYS = {
    "bucket_id": None,
    "folder_name": None,
    "show_versions": None,
    "recursive": None,
    "file_versions": FILE_VERSION_KEYS,
}

NOTIFICATION_RULES = {
    "bucket_id": None,
    "notification_rules": {
        "event_types": None,
        "is_enabled": None,
        "name": None,
        "object_name_prefix": None,
        "target_configuration": {
            "target_type": None,
            "url": None,
            "custom_headers": None,
            "hmac_sha256_signing_secret": None,
        },
    },
    "is_suspended": None,
    "suspension_reason": None,
}

BUCKET_KEYS = {
    "bucket_id": None,
    "bucket_name": None,
    "bucket_type": None,
    "bucket_info": None,
    "cors_rules": {
        "cors_rule_name": None,
        "allowed_origins": None,
        "allowed_operations": None,
        "max_age_seconds": None,
        "allowed_headers": None,
        "expose_headers": None,
    },
    "file_lock_configuration": {
        "is_file_lock_enabled": None,
        "default_retention": {
            "mode": None,
            "period": {
                "duration": None,
                "unit": None,
            },
        },
    },
    "default_server_side_encryption": BUCKET_SERVER_SIDE_ENCRYPTION,
    "lifecycle_rules": {
        "file_name_prefix": None,
        "days_from_hiding_to_deleting": None,
        "days_from_uploading_to_hiding": None,
    },
    "account_id": None,
    "options": None,
    "revision": None,
}
