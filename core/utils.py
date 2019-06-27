import os
import sys
import glob
import json
import requests
import time
import re
import ipaddress
import socket
import shutil
import subprocess

from pprint import pprint
from urllib.parse import quote, unquote
import urllib.parse
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
#################

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
ques = '{0}[?]{1} '.format(C, W)
bad = '{0}[-]{1} '.format(R, W)
good = '{0}[+]{1} '.format(G, W)

headers = {"User-Agent": "Osmedeus/v1.4", "Accept": "*/*",
           "Content-type": "application/json", "Connection": "close"}

# send request through Burp proxy for debug purpose
PROXY = {
    'http': 'http://127.0.0.1:8081',
    'https': 'http://127.0.0.1:8081'
}


def print_banner(text):
    print('{1}--~~~=[  {2}{0}{1} ]=~~~--'.format(text, G, C))


def print_target(text):
    print('{2}---<---<--{1}@{2} Target: {0} {1}@{2}-->--->---'.format(text, P, G))


def print_info(text):
    print(info + text)


def print_ques(text):
    print(ques + text)


def print_good(text):
    print(good + text)


def print_bad(text):
    print(bad + text)


def print_debug(text):
    print(G + "#" * 20 + GR)
    print(text)
    print("#" * 20)


def print_line():
    print(GR + '-' * 50)


def check_output(output):
    if str(output) != '' and str(output) != "None":
        print('{1}--==[ Check the output: {2}{0}{1}'.format(output, G, P))

#################


def url_encode(string_in):
    return urllib.parse.quote(string_in)


def url_decode(string_in):
    return urllib.parse.unquote(string_in)


def strip_slash(string_in):
    return string_in.replace('/', '_')


def replace_argument(options, cmd):
    for key, value in options.items():
        if key in cmd:
            cmd = cmd.replace('$' + str(key), str(value))
    return cmd


def make_directory(directory):
    if directory and not os.path.exists(directory):
        print_good('Make new directory: {0}'.format(directory))
        os.makedirs(directory)


# checking speed
def custom_speed(options):
    custom_speed = options.get('SLOW')
    if not custom_speed:
        return 'quick'

    if custom_speed != "None":
        # split at upper case
        raw_current_module = re.findall(r'[A-Z][^A-Z]*', options.get('CURRENT_MODULE'))
        current_module = [x.lower() for x in raw_current_module]

        if ',' in custom_speed:
            affected_modules = custom_speed.split(',')
            if any(elem in current_module for elem in affected_modules):
                print_good('Change speed of {0} module to slow'.format(
                    options.get('CURRENT_MODULE')))
                return 'slow'
        else:
            if custom_speed in current_module:
                print_good('Change speed of {0} module to slow'.format(
                    options.get('CURRENT_MODULE')))
                return 'slow'
            else:
                return 'quick'
    else:
        return 'slow'


# check if command run success or not
def cmd_exists(cmd):
    return subprocess.call("type " + cmd, shell=True,
        stdout=subprocess.PIPE, stderr=subprocess.PIPE) == 0


# check if program exist or command 
def is_installed(program_path=None, cmd=None):
    if cmd:
        return cmd_exists(cmd)
    program_path = os.path.normpath(program_path)
    if os.path.exists(program_path) and shutil.which(program_path):
        return True
    return False


def not_empty_file(filepath):
    fpath = os.path.normpath(filepath)
    return os.path.isfile(fpath) and os.path.getsize(fpath) > 0


# checking connection on port
def connection_check(target, port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    result = sock.connect_ex((target, int(port)))
    if result == 0:
        return True
    else:
        return False


def chunks(l, n):
    """Yield successive n-sized chunks from l."""
    for i in range(0, len(l), n):
        yield l[i:i + n]

############


'''
File Utils
'''


def list_files(folder, ext='xml', pattern=None):
    folder = os.path.normpath(folder)
    if pattern:
        return glob.glob(folder + '/{0}'.format(pattern))
    else:
        return glob.glob(folder + "/*." + ext)


def just_read(filename):
    filename = os.path.normpath(filename)
    if os.path.isfile(filename):
        with open(filename, 'r') as f:
            data = f.read()
        return data

    return False


def just_write(filename, data, is_json=False):
    filename = os.path.normpath(filename)
    try:
        print_good("Writing {0}".format(filename))
        if is_json:
            with open(filename, 'w+') as f:
                json.dump(data, f)
        else:
            with open(filename, 'w+') as f:
                f.write(data)
    except:
        print_bad("Writing fail: {0}".format(filename))
        return False


def just_append(filename, data):
    filename = os.path.normpath(filename)
    try:
        print_good("Append {0}".format(filename))
        with open(filename, 'a+') as f:
            f.write(data)
    except:
        print_bad("Append fail: {0}".format(filename))
        return False


def reading_json(filename):
    filename = os.path.normpath(filename)
    try:
        if os.path.isfile(filename):
            with open(filename, 'r') as f:
                main_json = json.load(f)
            return main_json
    except:
        print_bad("Reading fail: {0}".format(filename))

    return None

# just clean blank line and unique stuff
def clean_up(filename):
    filename = os.path.normpath(filename)
    if os.path.isfile(filename):
        with open(filename, 'r') as f:
            data = f.read().splitlines()

        final = []
        for item in data:
            if item != '':
                final.append(item.strip())
        final = set(final)

        just_write(filename, "\n".join(final))
        return filename

    return False


# just get main domain
def get_domain(string_in):
    parsed = urllib.parse.urlparse(string_in)
    domain = parsed.netloc if parsed.netloc else parsed.path
    return domain


# get IP of input string
def resolve_input(string_in):
    if valid_ip(string_in):
        return string_in
    else:
        try:
            ip = socket.gethostbyname(get_domain(string_in))
            return ip
        except:
            return False

    return False


# extract ip from file
def extract_ip(filename):
    data = just_read(filename)
    if data:
        lines = data.splitlines()
    else:
        return False

    ips = []
    for line in lines:
        if valid_ip(line):
            ips.append(line)
    return ips


# check if string is IP or not
def valid_ip(string_in):
    try:
        ipaddress.ip_interface(str(string_in).strip())
        return True
    except:
        return False
############

'''
Options utils
'''

def is_direct_mode(options, require_input=False):
    if options['MODULE'] != "None":
        print_good("Direct mode detect")
        if require_input:
            # input as a string
            if options.get('INPUT') and options.get('INPUT') != "None":
                return options.get('INPUT')

            # input as a file
            elif options.get('INPUT_LIST') and options.get('INPUT_LIST') != "None":
                if not_empty_file(options.get('INPUT_LIST')):
                    return options.get('INPUT_LIST')
                else:
                    print_bad(
                        "You you want to specific -i, -I options or file not found.")
                    sys.exit(-1)
            else:
                print_bad(
                    "You you want to specific -i, -I options or file not found.")
                sys.exit(-1)

        else:
            return True

    return False


def is_force(options, filename):
    if options.get('FORCE') != "False":
        return False

    if not_empty_file(filename):
        print_info(
            "Command is already done. use '-f' options to force rerun the command")
        return True
    return False

##################


'''
API Utils
'''


# get workspace name from options or direct string to prevent LFI
def get_workspace(options=None, workspace=None):
    if workspace:
        ws_name = os.path.basename(os.path.normpath(workspace))
        return ws_name

    elif options:
        ws_name = os.path.basename(
            os.path.normpath(options.get('STRIP_TARGET')))
    return ws_name


# adding times to waiting default is infinity times
def just_waiting(options, module_name, seconds=30, times=False):
    elapsed_time = 0
    if times:
        count = 0

    print_info('Waiting for {0} module'.format(module_name))
    while not checking_done(options, module=module_name):
        if not times:
            # just don't print this too much
            if ((elapsed_time / seconds) % 10) == 0:
                print_info('Waiting for {0} module'.format(module_name))
            time.sleep(seconds)

        if times:
            print_info(
                'Waiting for {0} module {1}/{2}'.format(module_name, str(count), str(times)))
            if count == int(times):
                print_bad(
                    "Something bad with {0} module but force to continue".format(module_name))
                force_done(options, module_name)
                break
            count += 1
            time.sleep(seconds)

        elapsed_time += seconds


# return True if the module was done and False if this modile not done
def resume(options, module):
    headers['Authorization'] = options['JWT']
    workspace = get_workspace(options=options)
    # force to run the module again
    if options.get('FORCE') != "False":
        return False

    try:
        # checking final report for the module
        url = options['REMOTE_API'] + \
            "/api/module/{0}".format(options['OUTPUT'])
        r = requests.get(url, verify=False, headers=headers)

        if r.status_code == 200:
            reports = json.loads(r.text).get('reports')
            for item in reports:
                if item.get('module') == module:
                    return not_empty_file(options['WORKSPACES'] + "/" + item.get('report'))

        # checking for each command of module
        url = options['REMOTE_API'] + "/api/{0}/routines?module=".format(workspace) + module
        r = requests.get(url, verify=False, headers=headers)
        if r.status_code == 200:
            routines = json.loads(r.text).get('routines')

        is_all_command_done = 0
        for item in routines:
            if not_empty_file(item.get('output_path')):
                is_all_command_done += 1

        if is_all_command_done == len(routines):
            return True
        else:
            return False
    except:
        return False

# checking if command was done or not? and return a json result
def checking_done(options, cmd=None, module=None, get_json=False):
    headers['Authorization'] = options['JWT']
    workspace = get_workspace(options=options)
    url = options['REMOTE_API'] + "/api/{0}/activities".format(workspace)

    if cmd:
        r = requests.post(url, verify=False, headers=headers, json={'cmd': cmd})
    if module:
        r = requests.post(url + "?module=" + module, verify=False, headers=headers, json={})

    if r.status_code == 401:
        if cmd:
            r = requests.post(url, verify=False, headers=headers, json={'cmd': cmd})
        if module:
            r = requests.post(url + "?module=" + module,
                              headers=headers, json={})

    if r.status_code == 200:
        commands = json.loads(r.text)
        for cmd in commands['commands']:
            if cmd['status'] == 'Running':
                return False if not get_json else commands

    return True if not get_json else commands


# force to update activities to all Done
def force_done(options, module):
    headers['Authorization'] = options['JWT']
    workspace = get_workspace(options=options)
    url = options['REMOTE_API'] + "/api/{0}/activities?module=".format(workspace) + module
    r = requests.put(url, verify=False, headers=headers)


def looping(options, cmd=None, module=None, times=5):
    headers['Authorization'] = options['JWT']
    workspace = get_workspace(options=options)
    url = options['REMOTE_API'] + "/api/{0}/activities".format(workspace)
    while times != 0:
        done = checking_done(options, cmd, module, url)
        if done:
            return
        times -= 1


# just for conclusion
def save_all_cmd(options, logfile, module=None):
    headers['Authorization'] = options['JWT']
    workspace = get_workspace(options=options)
    url = options['REMOTE_API'] + "/api/{0}/activities".format(workspace)

    if module:
        url += '?module=' + module

    r = requests.get(url, verify=False, headers=headers)

    with open(logfile, 'w+') as l:
        l.write(r.text)
    # commands = json.loads(r.text)['commands']


def get_jwt(options):
    username = options['USERNAME']
    password = options['PASSWORD']
    workspace = get_workspace(options=options)

    # set workspaces
    url = options['REMOTE_API'] + "/api/{0}/auth".format(workspace)
    data = {'username': username, 'password': password}
    r = requests.post(url, verify=False, headers=headers, json=data)

    if r.status_code == 200:
        if json.loads(r.text).get('access_token'):
            print_good("Authentication success on {0} workspace".format(workspace))
            token = "Bearer " + json.loads(r.text).get('access_token')
            return token
    return False


# set config file from remote
def set_config(options):
    url = options['REMOTE_API'] + "/api/config"
    # set workspaces
    data = {'options': options}
    r = requests.post(url, verify=False, headers=headers, json=data)
    return r


def just_shutdown_flask(options):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/shutdown"
    requests.post(url, verify=False, headers=headers)
