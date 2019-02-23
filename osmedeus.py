#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import os, sys, glob, socket, time
import argparse
from multiprocessing import Process
from pprint import pprint

from core import routine
from core import config
from core import execute
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
colors = [G,R,B,P,C,O,GR]


#############
# Osmedeus - One line to rude them all
#############

__author__ = '@j3ssiejjj'
__version__ = '1.0'


### Global stuff
current_path = os.path.dirname(os.path.realpath(__file__))
SPECIAL_ARGUMENT = {
    'TARGET' : 'example.com',
    'COMPANY' : 'example.com',
    'STRIP_TARGET' : 'example.com',
    'IP' : '1.2.3.4',
    'BURPSTATE' : '',
    'OUTPUT' : 'out.txt',
    'WORKSPACE' : current_path + '/workspaces',

    'PLUGINS_PATH' : current_path + '/plugins',
    'GO_PATH' : '~/go/bin',
    'DIRECTORY_FULL' : current_path + '/plugins/wordlists/dir-all.txt',
    'DOMAIN_FULL' : current_path + '/plugins/wordlists/all.txt',
    'DEFAULT_WORDLIST' : '',

    'GITHUB_API_KEY' : 'abc123poi456', # this isn't work api :D 
    'MORE' : '',
    'CWD' : os.path.dirname(os.path.realpath(__file__)),
}
###

options = {
    'target' : '',
    'targetlist' : '',
    'env' : SPECIAL_ARGUMENT,
    'speed' : 'quick',
    'DEBUG' : False
}


def flask_run():
    utils.print_banner("Staarting Flask API")
    os.system('python3 core/app.py')

def parsing_argument(args):
    
    p = Process(target=flask_run)
    p.start()

    #parsing agument
    if args.debug:
        options['DEBUG'] = True

    #parsing agument
    if args.git:
        options['env']['TARGET'] = args.git
        # git_routine(options)

    if args.burp:
        options['env']['BURPSTATE'] = args.burp

    if args.more:
        options['env']['MORE'] = args.more

    #choose speed to run this, default is quick
    if args.quick:
        options['speed'] = 'quick'
    if args.slow:
        options['speed'] = 'slow'


    if args.targetlist:
        options['targetlist'] = args.targetlist
        #check if target list file exist and loop throught the target
        if os.path.exists(options['targetlist']):
            with open(options['targetlist'], 'r+') as ts:
                targetlist = ts.read().splitlines()
            
            for target in targetlist:
                args.target = target
                single_target(args)
                print("{2}>++('> >++('{1}>{2} Target done: {0} {1}<{2}')++< <')++<".format(args.target, P, G))

    else:
        single_target(args)


def single_target(args):
    print('{2}---<---<--{1}@{2} Target: {0} {1}@{2}-->--->---'.format(args.target, P, G))
    if args.target:
        if args.output:
            options['env']['OUTPUT'] = args.output
        else:
            options['env']['OUTPUT'] = args.target

        #just loop in the for if the target list
        options['target'] = args.target
        options['env']['TARGET'] = args.target
        options['env']['STRIP_TARGET'] = args.target.replace('https://','').replace('http://','')
        if '/' in options['env']['STRIP_TARGET']:
            options['env']['STRIP_TARGET'] = options['env']['STRIP_TARGET'].split('/')[0]

        if args.workspace:
            if args.workspace[-1] == '/':
                options['env']['WORKSPACE'] = args.workspace + options['env']['STRIP_TARGET']
            else:
                options['env']['WORKSPACE'] = args.workspace + '/' + options['env']['STRIP_TARGET']
        else:
            options['env']['WORKSPACE'] = current_path + '/workspaces/' + options['env']['STRIP_TARGET']

        #create workspace folder for the target
        utils.make_directory(options['env']['WORKSPACE'])

        options['env']['COMPANY'] = args.target
        #checking for connection to target
        options['env']['IP'] = socket.gethostbyname(options['env']['TARGET'])


    #run specific task otherwise run the normal routine
    if args.module:
        module = args.module
        routine.specific(options, module)

    else:
        if options['DEBUG']:
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
    parser.add_argument('-m','--module' , action='store', dest='module', help='specific module to action')
    parser.add_argument('-t','--target' , action='store', dest='target', help='target')
    parser.add_argument('-T','--target_list' , action='store', dest='targetlist', help='list of target')
    parser.add_argument('-o','--output' , action='store', dest='output', help='output')
    parser.add_argument('-b','--burp' , action='store', dest='burp', help='burp http file')
    parser.add_argument('-g','--git' , action='store', dest='git', help='git repo to scan')
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
