#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import argparse
from pathlib import Path
import django

# Load default path of extensions folder
# WARNING: change this if you custom something
BASE_DIR = Path(os.path.dirname(
    os.path.dirname(os.path.abspath(__file__))))
sys.path.append(str(BASE_DIR.joinpath('server')))

# loading django setup
os.environ.setdefault("DJANGO_SETTINGS_MODULE", "rest.settings")
django.setup()

from api.models import *

from core import dbutils
from core import utils


# reload routine
def reload_routine():
    # load command from default workflow
    dbutils.internal_parse_commands()
    # create server config file
    dbutils.load_default_config(config_file='~/.osmedeus/server.conf')


def main(args):
    # workspace = args.workspace if args.workspace else None
    utils.print_block("Reload routine and config for server", tag='RUN')
    reload_routine()


parser = argparse.ArgumentParser(description="Reload routine for server")
parser.add_argument('-w', '--workspace', action='store',
                    dest='workspace', help='workspace')
args = parser.parse_args()
main(args)
