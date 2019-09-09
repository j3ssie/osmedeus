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


def clean_data_tables(workspace):
    utils.print_info('Clean Data Tables')
    if workspace:
        Activities.objects.filter(workspace=workspace).delete()
        Workspaces.objects.filter(workspace=workspace).delete()
        Summaries.objects.filter(workspace=workspace).delete()
        Reports.objects.filter(workspace=workspace).delete()
    else:
        Activities.objects.all().delete()
        Workspaces.objects.all().delete()
        Summaries.objects.all().delete()
        Reports.objects.all().delete()


def clean_stateless_tables():
    utils.print_info('Clean Stateless Tables')
    Configurations.objects.all().delete()
    Commands.objects.all().delete()
    ReportsSkeleton.objects.all().delete()
    Exploits.objects.all().delete()


# load default config
def load_default_config():
    # load command from default workflow
    dbutils.internal_parse_commands()
    # create server config file
    dbutils.load_default_config(config_file='~/.osmedeus/server.conf')
    # Configurations.objects.create(**record)
    utils.print_good("Load config success")


def main(args):
    workspace = args.workspace if args.workspace else None
    everything = args.all if args.all else None
    if workspace:
        clean_data_tables(workspace)

    if everything:
        clean_data_tables(None)
        clean_stateless_tables()
        load_default_config()


parser = argparse.ArgumentParser(description="Reset Database and Workspace")
parser.add_argument('-w', '--workspace', action='store',
                    dest='workspace', help='workspace')
parser.add_argument('-a', '--all', action='store',
                    help='Clean all')
args = parser.parse_args()
main(args)
