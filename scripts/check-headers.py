######################################################################
#
# File: scripts/check-headers.py
#
# Copyright 2021 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import sys
from glob import glob
from itertools import islice


try:
    pattern = sys.argv[1]
except IndexError:
    print('usage: python check-headers.py GLOB_PATTERN')
    sys.exit(2)

ignored = sys.argv[1:]

for file in glob(pattern, recursive=True):
    if file in ignored:
        continue
    with open(file) as fd:
        head = ''.join(islice(fd, 9))
        if 'All Rights Reserved' not in head:
            print(
                'Missing "All Rights Reserved" in the header in: {}'.format(file), file=sys.stderr
            )
            sys.exit(1)
        if file not in head:
            print('Wrong file name in the header in: {}'.format(file), file=sys.stderr)
            sys.exit(1)
