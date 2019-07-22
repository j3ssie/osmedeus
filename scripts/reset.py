#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import shutil
import argparse
from pathlib import Path
from configparser import ConfigParser, ExtendedInterpolation

BASE_DIR = Path(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
sys.path.append(str(BASE_DIR))

from core import utils


# change default config path if you custom it
def default_config(config_path=None):
    if not config_path:
        config_path = str(Path.home().joinpath('.osmedeus/config.conf'))
    utils.print_info("Detect config at: {0}".format(config_path))

    core_config = ConfigParser(interpolation=ExtendedInterpolation())
    core_config.read(config_path)
    sections = core_config.sections()
    options = {
        'CONFIG_PATH': os.path.abspath(config_path),
    }
    for sec in sections:
        for key in core_config[sec]:
            options[key.upper()] = core_config.get(sec, key)

    utils.print_info("Clean config.conf file")
    os.remove(config_path)

    return options

def delete_workspace(options):
    utils.print_info("Clean all workspace result")
    workspaces_path = os.path.normpath(options.get('WORKSPACES'))
    shutil.rmtree(workspaces_path)

def delete_storages(options):
    utils.print_info("Clean all storages result")
    storages_path = str(BASE_DIR.joinpath('core/rest/storages'))
    shutil.rmtree(storages_path)


def load_default_config(config_path):
    # change this to your config if you custom config
    options = default_config(config_path)
    delete_workspace(options)
    delete_storages(options)


def main():
    parser = argparse.ArgumentParser(
        description="Script to clean up db for Osmedeus")
    parser.add_argument('-c', '--config', action='store', dest='config',
                        help='Path your config.conf file (default: ~/.osmedeus/config.conf)')

    args = parser.parse_args()
    if len(sys.argv) == 1:
        parser.print_help()
        sys.exit(0)

    if args.config:
        load_default_config(args.config)
    else:
        load_default_config(None)


if __name__ == '__main__':
    main()
