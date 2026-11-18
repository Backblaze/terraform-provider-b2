######################################################################
#
# File: scripts/actions-changelog.py
#
# Copyright 2026 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

"""Repo-local replacements for the configured changelog Actions contracts."""

import argparse
from datetime import date
import json
import os
from pathlib import Path
import re
import sys
import uuid


VERSION = r'(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?'
HEADING = re.compile(r'## \[(Unreleased|' + VERSION + r')\](?: - (\d{4}-\d{2}-\d{2})(?: \[YANKED\])?)?')
KINDS = {'Added', 'Changed', 'Deprecated', 'Removed', 'Fixed', 'Security', 'Infrastructure', 'Updated'}
REFERENCE = re.compile(r'^\[([^]]+)\]:\s*https?://')


def entries(text):
    lines = text.splitlines()
    if not lines or not re.fullmatch(r'# \S.*', lines[0]):
        raise ValueError('Invalid changelog: the first line must be its title')
    result = []
    seen = set()
    previous = None
    kinds = set()
    for index, line in enumerate(lines):
        if line.startswith('# ') and index:
            raise ValueError('Invalid changelog: only one title is allowed')
        if line.startswith('## '):
            match = HEADING.fullmatch(line)
            if not match:
                raise ValueError(f'Invalid changelog heading on line {index + 1}')
            version = match[1]
            released = re.search(r' - (\d{4}-\d{2}-\d{2})', line)
            if version in seen or (version == 'Unreleased' and result):
                raise ValueError('Invalid changelog: duplicate or misplaced version')
            if version != 'Unreleased':
                if not released:
                    raise ValueError('Invalid changelog: release date is required')
                try:
                    date.fromisoformat(released[1])
                except ValueError as error:
                    raise ValueError('Invalid changelog: invalid release date') from error
                core = tuple(int(part) for part in re.split(r'[-+]', version)[0].split('.'))
                if previous is not None and core > previous and not re.search(r'alpha|rc|beta|dev', version, re.I):
                    raise ValueError('Invalid changelog: versions must be newest first')
                previous = core
            if index == 0 or lines[index - 1] != '' or index + 1 >= len(lines) or lines[index + 1] != '':
                raise ValueError('Invalid changelog: version headings need surrounding blank lines')
            if result:
                result[-1]['end'] = index
            result.append({'version': version, 'date': released[1] if released else '',
                           'start': index + 1, 'end': len(lines), 'heading': line})
            seen.add(version)
            kinds = set()
        elif line.startswith('### '):
            kind = line[4:]
            if not result or kind not in KINDS or kind in kinds:
                raise ValueError('Invalid changelog: invalid or duplicate change kind')
            if lines[index - 1] != '' or index + 1 >= len(lines) or not lines[index + 1]:
                raise ValueError('Invalid changelog: change headings need a blank line before and entries after')
            kinds.add(kind)
    if not result:
        raise ValueError('Invalid changelog: no version headings found')
    return result


def extract(text, version, notes_file):
    entry = next((entry for entry in entries(text) if entry['version'] == version), None)
    if entry is None:
        raise ValueError(f'Changelog version {version} not found')
    lines = text.splitlines()
    # Match changelog-reader's body: omit the version heading and reference
    # definitions, preserve the leading blank line, then resolve [name][] links.
    body_lines = lines[entry['start']:entry['end']]
    while body_lines and not body_lines[-1].strip():
        body_lines.pop()
    body = '\n'.join(line for line in body_lines if not REFERENCE.match(line))
    body = body.rstrip()
    registry = {REFERENCE.match(line)[1].lower(): line for line in lines if REFERENCE.match(line)}
    references = dict.fromkeys(name.lower() for name in re.findall(r'\[([^]]+)\]\[\]', body))
    for name in references:
        if name in registry:
            body += '\n' + registry[name]
    notes_file.write_text(body, encoding='utf-8')
    status = 'prereleased' if '-' in version else 'released'
    if '[YANKED]' in entry['heading']:
        status = 'yanked'
    if version == 'Unreleased':
        status = 'unreleased'
    if os.environ.get('GITHUB_OUTPUT'):
        delimiter = 'changelog_' + uuid.uuid4().hex
        with open(os.environ['GITHUB_OUTPUT'], 'a', encoding='utf-8') as output:
            output.write(f'version={version}\ndate={entry["date"]}\nstatus={status}\n')
            output.write(f'changes_file={notes_file.resolve()}\n')
            output.write(f'changes<<{delimiter}\n{body}\n{delimiter}\n')
    print(f'Extracted CHANGELOG.md [{version}] to {notes_file}')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest='command', required=True)
    read = commands.add_parser('extract')
    read.add_argument('version')
    read.add_argument('--notes-file', type=Path, required=True)
    validate = commands.add_parser('validate')
    validate.add_argument('--event-file', type=Path)
    validate.add_argument('--changed-files', type=Path)
    for command in (read, validate):
        command.add_argument('--path', type=Path, default=Path('CHANGELOG.md'))
    args = parser.parse_args()
    try:
        if args.command == 'validate' and args.event_file:
            event = json.loads(args.event_file.read_text(encoding='utf-8'))
            labels = {label['name'] for label in event['pull_request']['labels']}
            if '-changelog' in labels:
                print('Skipping changelog validation: -changelog label')
                return
            if args.changed_files is None or 'CHANGELOG.md' not in args.changed_files.read_text(encoding='utf-8').splitlines():
                raise ValueError('CHANGELOG.md must be modified (or apply the -changelog label)')
        text = args.path.read_text(encoding='utf-8')
        if args.command == 'extract':
            extract(text, args.version, args.notes_file)
        else:
            entries(text)
            # zattoo's package.json/monorepo release rules do not apply to this
            # Go provider. Its release version comes from the tag, checked by extract.
            print('CHANGELOG.md format and configured PR requirements: PASS')
    except (ValueError, KeyError, OSError) as error:
        parser.exit(1, str(error) + '\n')


if __name__ == '__main__':
    main()
