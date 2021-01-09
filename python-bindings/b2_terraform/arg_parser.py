######################################################################
#
# File: python-bindings/tf_provider_b2/arg_parser.py
#
# Copyright 2021 Backblaze Inc. All Rights Reserved.
#
# License https://www.backblaze.com/using_b2_code.html
#
######################################################################

import argparse


class ArgumentParser(argparse.ArgumentParser):
    def error(self, message):
        raise RuntimeError(message)
