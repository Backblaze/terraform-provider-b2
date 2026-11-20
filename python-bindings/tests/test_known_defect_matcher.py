######################################################################
#
# File: python-bindings/tests/test_known_defect_matcher.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

# Tests for the `known_defect` matcher in known_defect_plugin.py. Each case runs a tiny test module in a
# nested pytest session (pytester) that loads the same plugin the real suite uses.

import pytest

KNOWN = "@pytest.mark.known_defect(raises=AssertionError, match='the defect is here')"


@pytest.fixture
def run(pytester):
    pytester.makeini('[pytest]\nmarkers =\n    known_defect\n')
    pytester.makeconftest('from known_defect_plugin import *  # noqa: F403')

    def _run(source, *args):
        pytester.makepyfile(f'import pytest\n\n{source}')
        return pytester.runpytest(*args)

    return _run


def test_the_expected_failure_is_an_expected_failure(run):
    result = run(
        f"{KNOWN}\ndef test_x():\n    raise AssertionError('the defect is here: details')\n"
    )
    result.assert_outcomes(xfailed=1)


def test_an_unrelated_exception_type_fails(run):
    result = run(f"{KNOWN}\ndef test_x():\n    raise KeyError('the defect is here')\n")
    result.assert_outcomes(failed=1)


def test_the_expected_type_with_another_message_fails(run):
    result = run(f"{KNOWN}\ndef test_x():\n    raise AssertionError('some other assertion')\n")
    result.assert_outcomes(failed=1)


def test_an_unexpected_pass_fails(run):
    result = run(f'{KNOWN}\ndef test_x():\n    pass\n')
    result.assert_outcomes(failed=1)
    result.stdout.fnmatch_lines(['*XPASS(strict)*'])


def test_a_fixture_error_is_an_error_not_an_expected_failure(run):
    source = (
        "@pytest.fixture\ndef broken():\n    raise AssertionError('the defect is here')\n\n"
        f'{KNOWN}\ndef test_x(broken):\n    pass\n'
    )
    result = run(source)
    result.assert_outcomes(errors=1)


def test_runxfail_exposes_the_original_failure(run):
    result = run(
        f"{KNOWN}\ndef test_x():\n    raise AssertionError('the defect is here: details')\n",
        '--runxfail',
    )
    result.assert_outcomes(failed=1)
    result.stdout.fnmatch_lines(['*the defect is here: details*'])


def test_a_marker_without_an_expected_failure_is_rejected(run):
    result = run('@pytest.mark.known_defect\ndef test_x():\n    raise AssertionError()\n')
    assert result.ret != 0
    result.stderr.fnmatch_lines(['*known_defect needs raises=*match=*'])
