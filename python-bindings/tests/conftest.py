######################################################################
#
# File: python-bindings/tests/conftest.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import os
import sys

import pytest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from b2sdk.v3 import (  # noqa: E402
    B2Api,
    B2HttpApiConfig,
    DummyCache,
    InMemoryAccountInfo,
    RawSimulator,
)

from b2_terraform import provider_tool  # noqa: E402

# Offline fixtures for the python bindings: b2sdk's RawSimulator stands in for B2, so no credentials or
# network are needed. Install the dependencies with `pip install -r requirements.txt pytest`.
#
# Known defects: a test marked `@pytest.mark.known_defect` describes the correct behaviour of a defect that is
# still present. It is treated as a strict expected failure, so the default run stays green, and a test that
# starts passing fails the run ("XPASS(strict)") so that the marker gets removed. To run those tests as normal
# failing tests:
#
#     python -m pytest python-bindings/tests -m known_defect --runxfail
#
# When a defect is fixed, delete the `known_defect` marker from its test.


def pytest_collection_modifyitems(config, items):
    for item in items:
        if item.get_closest_marker('known_defect'):
            item.add_marker(pytest.mark.xfail(strict=True, reason='known defect'))


@pytest.fixture
def tool():
    """A ProviderTool whose api talks to the b2sdk simulator.

    DummyCache mirrors the real provider, which runs one fresh process per operation and so never
    has a bucket-name cache.
    """
    provider = provider_tool.ProviderTool()
    api = B2Api(
        account_info=InMemoryAccountInfo(),
        cache=DummyCache(),
        api_config=B2HttpApiConfig(_raw_api_class=RawSimulator),
    )
    account_id, master_key = api.session.raw_api.create_account()
    api.authorize_account(account_id, master_key)
    provider.__dict__['api'] = api  # ProviderTool.api is a cached_property
    return provider
