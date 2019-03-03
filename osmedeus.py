#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import os, sys, glob, socket, time
import argparse
from multiprocessing import Process
from pprint import pprint

from core import routine
from core import config
from core import execute
from core import slack
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
GR = '\033[1;37m'  # grays
colors = [G,R,B,P,C,O,GR]


#############
# Osmedeus - One line to rude them all
#############

__author__ = '@j3ssiejjj'
__version__ = '1.0'


### Global stuff
current_path = os.path.dirname(os.path.realpath(__file__))

def flask_run():
    utils.print_banner("Staarting Flask API")
    os.system('python3 core/app.py')

def parsing_argument(args):

    p = Process(target=flask_run)
    p.start()

    #parsing agument
    if args.config:
        config_path = args.config
        options = config.parsing_config(config_path, args)
            
    #wait for flask API start
    time.sleep(2)
    utils.set_config(options)

    if options['TARGET_LIST'] != "None":
        #check if target list file exist and loop throught the target
        if os.path.exists(options['TARGET_LIST']):
            with open(options['TARGET_LIST'], 'r+') as ts:
                targetlist = ts.read().splitlines()
            
            for target in targetlist:
                options['TARGET'] = target
                single_target(options)
                print(
                    "{2}>++('> >++('{1}>{2} Target done: {0} {1}<{2}')++< <')++<".format(options['TARGET'], P, G))

    else:
        single_target(options)


def single_target(options):
    print(
        '{2}---<---<--{1}@{2} Target: {0} {1}@{2}-->--->---'.format(options['TARGET'], P, G))
    slack.slack_seperate(options)
    #run specific task otherwise run the normal routine
    if options['MODULE'] != "None":
        module = options['MODULE']
        routine.specific(options, module)

    else:
        if options['DEBUG'] == "True":
            routine.debug(options)
        else:
            routine.normal(options)




def list_module():
    print(''' 
List module
===========
subdomain   - Scanning subdomain and subdomain takerover
portscan    - Screenshot and Scanning service for list of domain
brute       - Do brute force on service of target
vuln        - Scanning version of services and checking vulnerable service
git         - Scanning for git repo
burp        - Scanning for burp state
dirb        - Do directory search on the target
ip          - IP discovery on the target
headers     - Headers Scan on the target

        ''')
    sys.exit(0)

def update():
    execute.run1('git fetch --all && git reset --hard origin/master && ./install.sh')
    sys.exit(0)

def main():
    config.banner()
    parser = argparse.ArgumentParser(description="Collection tool for automatic pentesting")
    parser.add_argument('-c','--config' , action='store', dest='config', help='config file', default='core/config.conf')
    parser.add_argument('-m','--module' , action='store', dest='module', help='specific module to action')
    parser.add_argument('-t','--target' , action='store', dest='target', help='target')
    parser.add_argument('--company', action='store', dest='company', help='Company name')
    parser.add_argument('-b','--burp' , action='store', dest='burp', help='burp http file')
    parser.add_argument('-g','--git' , action='store', dest='git', help='git repo to scan')
    parser.add_argument('-T','--target_list' , action='store', dest='targetlist', help='list of target')
    parser.add_argument('-o','--output' , action='store', dest='output', help='output')
    parser.add_argument('-w','--workspace' , action='store', dest='workspace', help='Domain')
    parser.add_argument('--more' , action='store', dest='more', help='append more command for some tools')

    parser.add_argument('-M', '--list_module', action='store_true', help='List all module')
    parser.add_argument('-v', '--verbose', action='store_true', help='show verbose output')
    parser.add_argument('-f', '--force', action='store_true', help='force to run the module again if output exists')
    parser.add_argument('-q', '--quick', action='store_true', help='run this tool with quick routine', default=True)
    parser.add_argument('-s', '--slow', action='store_true', help='run this tool with slow routine', default=False)
    parser.add_argument('--mode', action='store_true', help='Choose mode to run normal routine(quick or slow)', default='quick')
    parser.add_argument('--update', action='store_true', help='update lastest from git')
    parser.add_argument('--debug', action='store_true', help='just for debug purpose')

    args = parser.parse_args()
    if len(sys.argv) == 1:
        list_module()
        sys.exit(0)

    if args.list_module:
        list_module()
    if args.update:
        update()

    parsing_argument(args)


if __name__ == '__main__':
    main()
