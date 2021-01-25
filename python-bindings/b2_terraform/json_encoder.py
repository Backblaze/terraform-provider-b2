######################################################################
#
# File: python-bindings/b2_terraform/json_encoder.py
#
# Copyright 2021 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import json


class B2ProviderJsonEncoder(json.JSONEncoder):
    """
    Makes it possible to serialize sets to json.
    """

    def default(self, obj):
        if isinstance(obj, set):
            return list(obj)
        return super(B2ProviderJsonEncoder, self).default(obj)
