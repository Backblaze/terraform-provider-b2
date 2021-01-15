######################################################################
#
# File: python-bindings/b2_terraform/arg_parser.py
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
