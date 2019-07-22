#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import subprocess
import shutil
import argparse
from pathlib import Path
from configparser import ConfigParser, ExtendedInterpolation

BASE_DIR = Path(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
sys.path.append(str(BASE_DIR))

from core import utils

# Console colors
W = '\033[1;0m'   # white
R = '\033[1;31m'  # red
G = '\033[1;32m'  # green
O = '\033[1;33m'  # orange
B = '\033[1;34m'  # blue
Y = '\033[1;93m'  # yellow
P = '\033[1;35m'  # purple
C = '\033[1;36m'  # cyan
GR = '\033[1;37m'  # gray
colors = [G, R, B, P, C, O, GR]

info = '{0}[*]{1} '.format(B, W)

ok = '{0}[{1} '.format(
    GR, G) + b'\xe2\x9c\x85'.decode('utf-8') + ' {0}] '.format(GR)

miss = '{0}[{1} '.format(
    GR, G) + b'\xe2\x9d\x8c'.decode('utf-8') + ' {0}] '.format(GR)


def print_ok(text):
    print(ok + text)


def print_miss(text):
    print(miss + text)


def print_info(text):
    print(info + text)


def not_empty_file(filepath):
    fpath = os.path.normpath(filepath)
    return os.path.isfile(fpath) and os.path.getsize(fpath) > 0

# check if command cane be execute or not
def cmd_exists(cmd):
    return subprocess.call("type " + cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE) == 0


# check if program exist or command
def is_installed(program_path=None, cmd=None):
    if shutil.which(program_path):
        return True
    return False


# change default config path if you custom it
def get_plugins_path(config_path=None, plugin_path=None):
    if not config_path:
        config_path = str(Path.home().joinpath('.osmedeus/config.conf'))
    if plugin_path:
        return Path(plugin_path)

    if os.path.isfile(os.path.normpath(config_path)):
        core_config = ConfigParser(interpolation=ExtendedInterpolation())
        core_config.read(config_path)
        plugin_path = Path(core_config.get('Enviroments', 'plugins_path'))
    else:
        plugin_path = BASE_DIR.joinpath('plugins')

    return plugin_path


def checking_files(plugin_path, files):
    for f in files:
        filepath = str(plugin_path.joinpath(f))
        if not not_empty_file(filepath):
            print_miss(" -- " + filepath)
        else:
            print_ok(" -- " + filepath)


def checking_program(plugin_path, command):
    for cmd in command:
        # command = str(plugin_path.joinpath(f))
        if not is_installed(program_path=cmd):
            print_miss(" -- " + cmd)
        else:
            print_ok(" -- " + cmd)


def checking_things(plugin_path):
    wordlist_files = [
        "wordlists/all.txt",
        "wordlists/shorts.txt",
        "wordlists/raft-large-directories.txt",
        "apps.json",
        "nmap-stuff/vulners.nse",
        "nmap-stuff/nmap_xml_parser.py",
        "nmap-stuff/masscan_xml_parser.py",
        "providers-data.csv",
    ]
    
    print_info("Checking for Wordlist")
    checking_files(plugin_path, wordlist_files)

    print_info("Checking for Plugins tools")
    go_files = [
        "go/amass",
        "go/subfinder",
        "go/gobuster",
        "go/aquatone",
        "go/webanalyze",
        "go/gowitness",
        "go/gitleaks",
        "dirsearch/dirsearch.py",
        "massdns/bin/massdns",
        "IPOsint/ip-osint.py",
    ]
    checking_files(plugin_path, go_files)

    builtin_programs = [
        "masscan",
        "nmap",
        "git",
        "git",
        "csvlook",
        "npm",
        "wfuzz",
    ]
    print_info("Checking for Built-in tools")
    checking_program(plugin_path, builtin_programs)


def load_default_config(config_path):
    # change this to your config if you custom config
    plugin_path = get_plugins_path(config_path)
    checking_things(plugin_path)


def main():
    parser = argparse.ArgumentParser(
        description="Script to clean up db for Osmedeus")
    parser.add_argument('-c', '--config', action='store', dest='config',
                        help='Path your config.conf file (default: ~/.osmedeus/config.conf)')
    parser.add_argument('-p', '--plugin', action='store', dest='plugin',
                        help='Path your plugin')

    args = parser.parse_args()
    # if len(sys.argv) == 1:
    #     parser.print_help()
    #     sys.exit(0)

    if args.config:
        load_default_config(args.config)
    else:
        load_default_config(None)


if __name__ == '__main__':
    main()
