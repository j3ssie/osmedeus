import os
import socket

from Osmedeus.core import execute
from Osmedeus.core import utils

class Healcheck(object):
    """docstring for Healcheck"""

    def __init__(self, options):
        utils.print_banner('Starting Healcheck')
        self.options = options
        self.healthcheck = True

    def checking(self):
        self.command_check()
        self.program_check()
        self.file_check()
        return self.healthcheck

    def file_check(self):
        files = [
            "$PLUGINS_PATH/wordlists/all.txt",
            "$PLUGINS_PATH/wordlists/shorts.txt",
            "$PLUGINS_PATH/wordlists/raft-large-directories.txt",
            "$PLUGINS_PATH/apps.json",
            "$PLUGINS_PATH/nmap-stuff/vulners.nse",
            "$PLUGINS_PATH/nmap-stuff/nmap_xml_parser.py",
            "$PLUGINS_PATH/nmap-stuff/masscan_xml_parser.py",
            "$PLUGINS_PATH/providers-data.csv",
        ]

        for f in files:
            filepath = utils.replace_argument(self.options, f)
            if not utils.not_empty_file(filepath):
                utils.print_bad("Bad -- " + filepath)
                self.healthcheck = False
            else:
                utils.print_good("Good -- " + filepath)

    def program_check(self):
        # check program
        programs = [
            "$GO_PATH/amass",
            "$GO_PATH/subfinder",
            "$GO_PATH/gobuster",
            "$GO_PATH/aquatone",
            "$GO_PATH/webanalyze",
            "$GO_PATH/gowitness",
            "$GO_PATH/gitleaks",
        ]

        for p in programs:
            program = utils.replace_argument(self.options, p)
            if not self.installed(program=program):
                utils.print_bad("Bad -- " + program)
                self.healthcheck = False
            else:
                utils.print_good("Good -- " + program)

    def command_check(self):
        # base command
        commands = [
            "masscan -h",
            "nmap -h",
            "git -h",
            "git -h",
            "csvlook -h",
            "npm -h",
            "wfuzz -h",
            "$PLUGINS_PATH/dirsearch/dirsearch.py -h"
            "$PLUGINS_PATH/massdns/bin/massdns -h"
            "$PLUGINS_PATH/IPOsint/ip-osint.py -h"
        ]

        for c in commands:
            command = utils.replace_argument(self.options, c)
            if not self.installed(cmd=command):
                utils.print_bad("Bad -- " + command)
                self.healthcheck = False
            else:
                utils.print_good("Good -- " + command)

    def installed(self, program=None, cmd=None):
        if program:
            return utils.is_installed(program_path=program)
        elif cmd:
            return utils.is_installed(cmd=cmd)
