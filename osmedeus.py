#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import os
import sys
import time
import argparse
from multiprocessing import Process

from core import routine
from core import config
from core import slack
from core import utils
from core import report


#############
# Osmedeus - One line to rude them all
#############

__author__ = '@j3ssiejjj'
__version__ = '1.4'


# run Flask API as another process
def flask_run():
    utils.print_banner("Starting Flask API")
    os.system('python3 core/app.py')


def parsing_argument(args):
    if not args.client and not args.report:
        if not utils.connection_check('127.0.0.1', 5000):
            p = Process(target=flask_run)
            p.start()
            # wait for flask API start
            time.sleep(2)
        else:
            utils.print_info("Look like Flask API already ran")

    # parsing agument
    if args.config:
        options = config.parsing_config(args.config, args)
    else:
        options = config.parsing_config(None, args)

    # report mode
    if args.report:
        report.parsing_report(options)
        sys.exit(0)

    if options['TARGET_LIST'] != "None":
        # check if target list file exist and loop throught the target
        targetlist = utils.just_read(options.get('TARGET_LIST'))
        if targetlist:
            for target in targetlist.splitlines():
                options['TARGET'] = target
                single_target(options)
                utils.print_target(options.get('TARGET'))
    else:
        single_target(options)


def single_target(options):
    try:
        utils.set_config(options)
        options['JWT'] = utils.get_jwt(options)
    except Exception:
        utils.print_bad(
            "Fail to set config, Something went wrong with Flask API !")
        utils.print_info(
            "Visit this page for common issue: https://github.com/j3ssie/Osmedeus/wiki/Common-Issues")
        sys.exit(-1)

    if not (options['JWT'] and options['JWT'] != "None"):
        utils.print_bad("Can't login to get JWT")
        sys.exit(-1)

    utils.print_target(options.get('TARGET'))

    # just disable slack noti in debug mode
    if options['DEBUG'] != "True":
        slack.slack_seperate(options)

    # run specific task otherwise run the normal routine
    if options['MODULE'] != "None":
        module = options['MODULE']
        routine.specific(options, module)

    else:
        if options['DEBUG'] == "True":
            routine.debug(options)
        else:
            routine.normal(options)


def main():
    config.banner(__version__, __author__)
    parser = argparse.ArgumentParser(
        description="One line to rude them all", add_help=False)

    parser.add_argument('-c', '--config', action='store', dest='config',
                        help='config file')
    parser.add_argument('-m', '--module', action='store',
                        dest='module', help='specific module to action')

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

    parser.add_argument('-f', '--force', action='store_true',
                        help='force to run the module again if output exists')

    # api options
    parser.add_argument('--remote', action='store', dest='remote', default="https://127.0.0.1:5000",
                        help='remote address for API')
    parser.add_argument('--auth', action='store', dest='auth',
                        help='Specify authentication e.g: --auth="username:password" ')
    parser.add_argument('--proxy', action='store', dest='proxy',
                        help='Specify proxy --proxy="type://host:port" e.g: --proxy="socks4://127.0.0.1:9050" ')
    parser.add_argument('--company', action='store',
                        dest='company', help='Company name')
    parser.add_argument('-b', '--burp', action='store',
                        dest='burp', help='burp http file')
    parser.add_argument('-g', '--git', action='store',
                        dest='git', help='git repo to scan')
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
    parser.add_argument('-h', '--help', dest='help',
                        action='store_true', help='Display help messaage')

    args = parser.parse_args()
    if len(sys.argv) == 1:
        config.custom_help()
        sys.exit(0)

    if args.help:
        config.custom_help()

    if args.list_module:
        config.list_module()
    if args.update:
        config.update()

    parsing_argument(args)


if __name__ == '__main__':
    main()
