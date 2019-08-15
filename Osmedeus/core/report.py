import os
import sys
from pathlib import Path
from tabulate import tabulate
import requests
import urllib3

from Osmedeus.core import utils
from Osmedeus.resources import *

sys.path.append(os.path.dirname(os.path.realpath(__file__)))
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

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

# Global stuff
headers = {"User-Agent": "Osmedeus/v1.5", "Accept": "*/*",
           "Content-type": "application/json", "Connection": "close"}


# checking result locally
def local_get_report(options):
    command_path = str(RESOURCES_PATH.joinpath('rest/commands.json'))
    commands = utils.reading_json(command_path)
    # create skeleton dict
    final_reports = []
    for key in commands.keys():
        final_reports.append({
            "module": key,
            "reports": []
        })
    # get workspace name
    ws_name = options.get('TARGET')

    for k in commands.keys():
        if "report" in commands[k].keys():
            report = utils.replace_argument(
                options, commands[k].get("report"))
            # @TODO refactor this later
            if type(report) == str:
                if utils.not_empty_file(report):
                    report_path = report.replace(
                        options.get('WORKSPACE'), ws_name)

                    report_item = {
                        "path": report_path,
                        "type": "html",
                    }
                    for i in range(len(final_reports)):
                        if final_reports[i].get('module') == k:
                            final_reports[i]["reports"].append(
                                report_item)

            elif type(report) == list:
                for item in report:
                    report_path = utils.replace_argument(
                        options, item.get("path"))
                    if utils.not_empty_file(report_path):
                        report_path = report_path.replace(
                            options.get('WORKSPACE'), ws_name)

                        report_item = {
                            "path": report_path,
                            "type": item.get("type"),
                        }
                        for i in range(len(final_reports)):
                            if final_reports[i].get('module') == k:
                                final_reports[i]["reports"].append(
                                    report_item)

    # just clean up
    clean_reports = []
    for i in range(len(final_reports)):
        if final_reports[i].get('reports'):
            clean_reports.append(final_reports[i])
    return {'reports': clean_reports}


# just sender
def sender(url, headers, data, method='POST'):
    if method == 'POST':
        r = requests.post(url, verify=False, headers=headers, json=data)
    elif method == 'GET':
        r = requests.get(url, verify=False, headers=headers)

    return r


# list all workspace
def workspace_list(options):
    global headers
    try:
        url = options['REMOTE_API'] + "/api/workspaces"
        r = sender(url, headers=headers, data={}, method='GET')
        if r:
            raw_workspaces = r.json().get('workspaces')
    except Exception:
        raw_workspaces = os.listdir(options.get('WORKSPACES'))

    workspaces = [[x] for x in raw_workspaces]
    headers = ["Available workspace"]
    print(tabulate(workspaces, headers, tablefmt="grid"))

    return workspaces


# just beautiful print
def quick_banner(module, filename):
    # content = [[filename]]
    headers = [module, filename]
    return tabulate([], headers, tablefmt="grid")


# strip out some modules for full report
def check_module(options, module):
    if options.get('MODULE') and options.get('MODULE') != "None":
        if ',' in options.get('MODULE'):
            match_module = options.get('MODULE').split(',')
            if any(elem.strip() in module.lower() for elem in match_module):
                return True
        elif options.get('MODULE').lower() in module.lower():
            return True
    elif options.get('MODULE') == "None":
        return True
    else:
        return False


# just reading
def read_report(report_path, force=False):
    if force:
        output = utils.just_read(report_path)
    else:
        if report_path.endswith('.html') or report_path.endswith('.xml'):
            output = report_path
        else:
            output = utils.just_read(report_path)

    if output:
        print(output)


# path report
def path_report(options, raw_reports):
    head = ['Module', 'Report Path']
    contents = []
    for item in raw_reports:
        for report in item.get('reports'):
            report_path = os.path.join(options.get(
                'WORKSPACES'), report.get('path'))
            contents.append([item.get('module'), report_path])

    print(tabulate(contents, head, tablefmt="grid"))


# short report
def short_report(options):
    global headers
    workspace = utils.get_workspace(options=options)
    try:
        url = options['REMOTE_API'] + "/api/module/{0}".format(workspace)
        r = sender(url, headers=headers, data={}, method='GET')
        if r:
            raw_reports = r.json().get('reports', None)
    except Exception:
        raw_reports = None

    if not raw_reports:
        raw_reports = local_get_report(options).get('reports')
        if not raw_reports:
            utils.print_bad(
                "Can't get log file for {0} workspace".format(options.get('TARGET')))
            return None

    # only print out the path
    if options.get('REPORT') == 'path':
        path_report(options, raw_reports)
        return None

    for item in raw_reports:
        for report in item.get('reports'):
            report_path = os.path.join(options.get('WORKSPACES'), report.get('path'))
            utils.print_info(item.get('module') + ": " + report_path)

            # checking if get specific module or not
            if check_module(options, item.get('module')):
                read_report(report_path)
            elif options.get('MODULE') == "None":
                read_report(report_path)
            utils.print_line()


# full report
def full_report(options):
    global headers
    try:
        workspace = utils.get_workspace(options=options)
        url = options['REMOTE_API'] + "/api/{0}/activities".format(workspace)
        r = sender(url, headers=headers, data={}, method='GET')
        if r:
            raw_reports = r.json()
    except Exception:
        log_path = Path(options.get('WORKSPACE')).joinpath('log.json')
        raw_reports = utils.reading_json(log_path)

    if not raw_reports:
        utils.print_bad(
            "Can't get log file for {0} workspace".format(options.get('TARGET')))
        return None

    modules = list(raw_reports.keys())
    for module in modules:
        if check_module(options, module):
            reports = raw_reports.get(module)
            utils.print_banner(module)
            for report in reports:
                cmd = report.get('cmd')
                utils.print_info("Command Executed: {0}\n".format(cmd))
                output_path = report.get('output_path')
                std_path = report.get('std_path')

                if 'raw' in options.get('REPORT').lower():
                    read_report(std_path)
                elif 'full' in options.get('REPORT').lower():
                    read_report(output_path)
                utils.print_line()


# summary report
def summary_report(options):
    global headers
    workspace = utils.get_workspace(options=options)
    try:
        url = options['REMOTE_API'] + "/api/workspace/{0}".format(workspace)
        r = sender(url, headers=headers, data={}, method='GET')
        if r:
            main_json = r.json()
    except Exception:
        main_json_path = Path(options.get('WORKSPACE')).joinpath(
            '{0}.json'.format(workspace))
        main_json = utils.reading_json(main_json_path)

    if not main_json:
        utils.print_bad(
            "Can't get log file for {0} workspace".format(options.get('TARGET')))
        return None

    subdomains = main_json.get('Subdomains')

    head = ['Domain', 'IP', 'Technologies', 'Ports']

    contents = []

    for element in subdomains:
        item = [
            element.get('Domain'),
            element.get('IP'),
            "\n".join(element.get('Technology')),
            ",".join(element.get('Ports')),
        ]
        contents.append(item)

    print(tabulate(contents, head, tablefmt="grid"))


# select report mode
def parsing_report(options):
    global headers
    mode = options.get('REPORT').lower()
    modes = ['full', 'raw', 'short', 'raw_html',
             'list', 'ls', 'path',
             'summary', 'sum'
             ]

    if mode not in modes:
        report_help()
        return None
    if not utils.connection_check('127.0.0.1', 5000):
        utils.print_info(
            "Look like API not turn on, reading from local log instead.")
    else:
        options['JWT'] = utils.get_jwt(options)
        headers['Authorization'] = options['JWT']

    # mapping mode
    if 'list' in mode or 'ls' in mode:
        workspace_list(options)
    elif 'path' in mode:
        short_report(options)
    elif 'short' in mode:
        short_report(options)
    elif 'full' in mode:
        full_report(options)
    elif 'raw' in mode:
        full_report(options)
    elif 'sum' in mode:
        summary_report(options)


# print report help message
def report_help():
    print('''{1}
{1}[{0}Report Mode{1}]{1}
===================
sum         - Summary report
list        - List avalible workspace
short       - Only print final output of each module
full        - Print all output of each module
path        - Only print final path of each module
raw         - Print all stdout of each module
html        - Export to html

{1}[{0}Filter module{1}]{1}
===================
subdomain, recon, assetfinding
takeover, screenshot
portscan, dirbrute, vulnscan
gitscan, cors, ipspace, sslscan, headers


{1}[{0}Report Usage{1}]{1}
===================
./osemdeus.py --report <mode> -t <workspace> [-m <module>]

{1}[{0}Example Commands{1}]{1}
===================
./osemdeus.py -t example.com --report list
./osemdeus.py -t example.com --report sum
./osemdeus.py -t example.com --report path
./osemdeus.py -t example.com --report short
./osemdeus.py -t example.com -m subdomain --report short
./osemdeus.py -t example.com -m subdomain, portscan --report short
./osemdeus.py -t example.com -m subdomain, portscan --report full
    '''.format(G, GR, B))
