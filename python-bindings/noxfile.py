######################################################################
#
# File: noxfile.py
#
# Copyright 2020 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import os
import platform
from glob import glob

import nox

CI = os.environ.get('CI') is not None

nox.options.reuse_existing_virtualenvs = True
nox.options.sessions = []

# In CI, use Python interpreter provided by GitHub Actions
if CI:
    nox.options.force_venv_backend = 'none'


@nox.session(python='3.9')
def bundle(session):
    """Bundle the distribution."""
    session.install('-e', '.', 'pyinstaller', 'staticx')
    session.run('rm', '-rf', 'build', 'dist', 'py-terraform-provider-b2.egg-info', external=True)
    session.run('pyinstaller', '--onefile', 'py-terraform-provider-b2.spec')

    # Set outputs for GitHub Actions
    if CI:
        asset_path = glob('dist/*')[0]
        print('::set-output name=asset_path::', asset_path, sep='')

        name, ext = os.path.splitext(os.path.basename(asset_path))
        system = platform.system().lower()
        asset_name = '{}-{}{}'.format(name, system, ext)
        print('::set-output name=asset_name::', asset_name, sep='')
