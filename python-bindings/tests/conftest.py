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

from b2_simulator import KEY, KEY_ID, B2Simulator
from b2sdk.v3 import (
    B2Api,
    DummyCache,
    InMemoryAccountInfo,
)
from known_defect_plugin import (  # noqa: F401
    pytest_collection_modifyitems,
    pytest_runtest_makereport,
)

from b2_terraform import provider_tool

# Offline fixtures for the python bindings: a standalone b2-simulator (backblaze-labs/b2-simulator, v0.6.0 or
# later) stands in for B2 on loopback, so no credentials are needed. Set B2_SIMULATOR_DIR to its checkout and
# have Node.js on PATH; tests that use the `tool` fixture fail with a clear error without it. Install the
# dependencies with `pip install -r requirements.txt pytest`.
#
# Known defects: a test marked `@pytest.mark.known_defect(raises=<exception type>, match='<regex>')` describes
# the correct behaviour of a defect that is still present. known_defect_plugin.py reports it as an expected
# failure only when it fails with that exact exception and message, so the default run stays green. Any other
# failure or fixture error fails the run, and a test that starts passing fails it too ("XPASS(strict)") so that
# the marker gets removed. To run those tests as normal failing tests:
#
#     python -m pytest python-bindings/tests -m known_defect --runxfail
#
# When a defect is fixed, delete the `known_defect` marker from its test.


@pytest.fixture(scope='session')
def b2_simulator():
    """One b2-simulator for the session; the `tool` fixture resets it before each test."""
    simulator = B2Simulator()
    yield simulator
    simulator.stop()


@pytest.fixture
def tool(b2_simulator):
    """A ProviderTool whose api talks to an empty b2-simulator.

    DummyCache mirrors the real provider, which runs one fresh process per operation and so never
    has a bucket-name cache.
    """
    b2_simulator.reset()
    provider = provider_tool.ProviderTool()
    api = B2Api(account_info=InMemoryAccountInfo(), cache=DummyCache())
    api.authorize_account(KEY_ID, KEY, realm=b2_simulator.api_url)
    provider.__dict__['api'] = api  # ProviderTool.api is a cached_property
    return provider
