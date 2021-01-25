######################################################################
#
# File: scripts/check-gofmt.py
#
# Copyright 2021 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import subprocess
import sys
from glob import glob


try:
    pattern = sys.argv[1]
except IndexError:
    print('usage: python check-gofmt.py GLOB_PATTERN')
    sys.exit(2)

ignored = sys.argv[1:]

for file in glob(pattern, recursive=True):
    if file in ignored:
        continue
    output = subprocess.run(['gofmt', '-l', file], capture_output=True).stdout.strip()
    if output:
        print('Go formatter needs running on the file: {}'.format(file), file=sys.stderr)
        sys.exit(1)
