######################################################################
#
# File: python-bindings/tests/known_defect_plugin.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

# Matcher for `@pytest.mark.known_defect(raises=<exception type>, match='<regex>')`.
#
# A marked test describes the correct behaviour of a defect that is still present. It is reported as an
# expected failure only when its call phase fails with an exception of type `raises` whose text matches the
# regex `match` (the exception that the defect produces today). Everything else is reported normally:
#   - a different exception, or the same type with another message: failed;
#   - an error while setting up a fixture: error (it is not the defect);
#   - the test passes: failed with "XPASS(strict)", so that the marker gets removed once the defect is fixed;
#   - `--runxfail`: the matcher is off and the original failure is shown.

import re

import pytest

MARKER = 'known_defect'


def _expected(item):
    marker = item.get_closest_marker(MARKER)
    if marker is None:
        return None
    exc_type = marker.kwargs.get('raises')
    pattern = marker.kwargs.get('match')
    if not (isinstance(exc_type, type) and issubclass(exc_type, BaseException) and pattern):
        raise pytest.UsageError(
            f'{item.nodeid}: known_defect needs raises=<exception type> and match=<regex> that '
            'identify the failure the defect produces today'
        )
    return exc_type, pattern, marker.kwargs.get('reason', 'known defect')


def pytest_collection_modifyitems(config, items):
    for item in items:
        _expected(item)  # reject an incomplete marker at collection time


@pytest.hookimpl(hookwrapper=True)
def pytest_runtest_makereport(item, call):
    outcome = yield
    report = outcome.get_result()
    expected = _expected(item)
    if expected is None or item.config.option.runxfail or call.when != 'call':
        return
    exc_type, pattern, reason = expected
    if call.excinfo is None:
        report.outcome = 'failed'
        report.longrepr = f'[XPASS(strict)] known defect no longer reproduces ({reason}); remove its known_defect marker'
    elif call.excinfo.errisinstance(exc_type) and re.search(pattern, str(call.excinfo.value)):
        report.outcome = 'skipped'
        report.wasxfail = reason
