#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import os
import sys
import time
import argparse
from multiprocessing import Process

from lib.core import utils
from lib.noti import slack
from lib.client import config
from lib.sender import auth
from lib.sender import initial
from lib.mode import routine


#############
# Osmedeus - One line to rude them all
#############

__author__ = '@j3ssiejjj'
__version__ = '2.0'


# run Django API as another process
def start_server():
    utils.print_banner("Starting Django API")
    os.system('python3 server/manage.py runserver')


def parsing_argument(args):
    # parsing agument
    options = config.parsing_config(args)

    # Start Django API if it's not
    if not args.client:
        if not utils.connection_check('127.0.0.1', 8000):
            utils.print_line()
            p = Process(target=start_server)
            p.start()
            # wait for Django API start
            time.sleep(3)
            utils.print_line()
        else:
            utils.print_info("Look like Django API already ran")
    options = auth.login(options)
    single_target(options)


def single_target(options):
    if not options or not (options['JWT'] and options['JWT'] != "None"):
        utils.print_bad("Can't login to get JWT")
        sys.exit(-1)

    # don't create new workspace in report mode
    if options.get('mode') != 'report':
        options = initial.init_workspace(options)
    # run specific task otherwise run the normal routine
    routine.routine_handle(options)


def main():
    config.banner(__version__, __author__)
    parser = argparse.ArgumentParser(
        description="One line to rude them all")

    parser.add_argument('-c', '--config', action='store', dest='config_path',
                        help='config file')
    parser.add_argument('-m', '--modules', action='store',
                        dest='modules', help='specific modules to action')

    # input
    parser.add_argument('-i', '--input', action='store',
                        dest='input', help='input for specific module')
    parser.add_argument('-I', '--input_list', action='store',
                        dest='inputlist', help='input file for specific module')
    parser.add_argument('-t', '--target', action='store',
                        dest='target', help='target')
    parser.add_argument('-T', '--target_list', action='store',
                        dest='targetlist', help='list of target')
    # report
    parser.add_argument('-r', '--report', action='store',
                        dest='report', help='report mode')
    # workspace
    parser.add_argument('-w', '--workspace', action='store',
                        dest='workspace', help='Domain')
    # more options on routine
    parser.add_argument('-s', '--slow', action='store',
                        help='run this tool with slow routine')

    parser.add_argument('-f', '--forced', action='store_true',
                        help='force to run the module again if output exists')

    # api options
    parser.add_argument('--remote', action='store', dest='remote', help='remote address for API')
    parser.add_argument('--auth', action='store', dest='auth',
                        help='Specify authentication e.g: --auth="username:password" ')
    parser.add_argument('--proxy', action='store', dest='proxy',
                        help='Specify proxy --proxy="type://host:port" e.g: --proxy="socks4://127.0.0.1:9050" ')

    parser.add_argument('--company', action='store',
                        dest='company', help='Company name')

    parser.add_argument('-o', '--output', action='store',
                        dest='output', help='output')

    parser.add_argument('-M', '--list_module',
                        action='store_true', help='List all module')
    parser.add_argument('-v', '--verbose', action='store_true',
                        help='show verbose output')

    parser.add_argument('--update', action='store_true',
                        help='update lastest from git')

    parser.add_argument('--proxy_file', action='store', dest='proxy_file',
                        help='Specify proxychains config file --proxy_file=proxychains.conf')

    parser.add_argument('--client', action='store_true',
                        help='just run client stuff in case you ran the flask server before')
    parser.add_argument('--debug', action='store_true',
                        help='just for debug purpose')

    parser.add_argument('--slack', action='store_true',
                        help='Turn on slack notification')

    parser.add_argument('--reset', action='store_true',
                        help='Clean configurations')
    parser.add_argument('-hh', '--helps', dest='helps',
                        action='store_true', help='Display more help message')

    args = parser.parse_args()
    if len(sys.argv) == 1:
        config.custom_help()
        sys.exit(0)

    if args.helps:
        config.custom_help()

    if args.list_module:
        config.list_module()
    if args.update:
        config.update()

    parsing_argument(args)


if __name__ == '__main__':
    main()
