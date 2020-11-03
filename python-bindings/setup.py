######################################################################
#
# File: setup.py
#
# Copyright 2019 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

from setuptools import find_packages, setup

with open('requirements.txt', encoding='utf-8') as f:
    requirements = f.read().splitlines()

setup(
    name='py-terraform-provider-b2',
    version='0.1.0',
    description='Python bindings for Backblaze B2 Terrafrom Provider',
    long_description_content_type='text/markdown',
    url='https://github.com/Backblaze/terraform-provider-b2',
    author='Backblaze, Inc.',
    author_email='support@backblaze.com',
    license='MIT',
    classifiers=[
        'Development Status :: 5 - Production/Stable',
        'Intended Audience :: Developers',
        'Topic :: Software Development :: Libraries',
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.5',
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Programming Language :: Python :: 3.9',
    ],
    keywords='backblaze b2 cloud storage',
    packages=find_packages(exclude=['test']),
    dependency_links=[],
    install_requires=requirements,
    package_data={'py-terraform-provider-b2': ['requirements.txt', 'LICENSE']},
    entry_points={
        'console_scripts': ['py-terraform-provider-b2=tf_provider_b2.__main__:main'],
    },
)
