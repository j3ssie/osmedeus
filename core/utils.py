import os
import sys
import json
import requests
import time
from pprint import pprint
from urllib.parse import quote, unquote
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

headers = {"User-Agent": "Osmedeus/v1.2", "Accept": "*/*",
           "Content-type": "application/json", "Connection": "close"}

#send request through Burp proxy for debug purpose
PROXY = {
    'http': 'http://127.0.0.1:8081',
    'https': 'http://127.0.0.1:8081'
}


def print_banner(text):
	print('{1}--~~~=:>[ {2}{0}{1} ]>'.format(text, G, C))


def print_info(text):
	print(info + text)


def print_ques(text):
	print(ques + text)


def print_good(text):
	print(good + text)


def print_bad(text):
	print(bad + text)


def check_output(output):
    if str(output) != '' and str(output) != "None":
        print('{1}--==[ Check the output: {2}{0}{1}'.format(output, G, P))

#################


def replace_argument(options, cmd):
	# for key,value in options['env'].items():
	for key, value in options.items():
		if key in cmd:
			cmd = cmd.replace('$' + str(key), str(value))
	return cmd


def make_directory(directory):
	if not os.path.exists(directory):
		print_good('Make new directory: {0}'.format(directory))
		os.makedirs(directory)


def not_empty_file(fpath):
	return os.path.isfile(fpath) and os.path.getsize(fpath) > 0

#checking connection


def connection_check(target):
	return True


def chunks(l, n):
    """Yield successive n-sized chunks from l."""
    for i in range(0, len(l), n):
        yield l[i:i + n]


def just_read(filename):
    if os.path.isfile(filename):
        with open(filename, 'r') as f:
            data = f.read()
        return data

    return False


def just_write(filename, data, is_json=False):
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


def reading_json(filename):
    try:
        if os.path.isfile(filename):
            with open(filename, 'r') as f:
                main_json = json.load(f)
            return main_json
    except:
        print_bad("Reading fail: {0}".format(filename))

    return None

#adding times to waiting default is infinity times


def just_waiting(options, module_name, seconds=30, times=False):
    if times:
        count = 0

    while not checking_done(options, module=module_name):
        if not times:
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

###


def is_direct_mode(options, require_input=False):
    if options['MODULE'] != "None":
        print_good("Direct mode detect")
        if require_input:
            if not_empty_file(options['INPUT']):
                return options['INPUT']
            else:
                print_bad("You you want to specific -i options")
                sys.exit(-1)
        else:
            return True

    return False

#return True if the module was done and False if this modile not done 
def resume(options, module):
    headers['Authorization'] = options['JWT']

    #force to run the module again
    if options.get('FORCE') != "False":
        return False

    try:
        #checking final report for the module
        url = options['REMOTE_API'] + "/api/module/{0}".format(options['OUTPUT']) 
        r = requests.get(url, verify=False, headers=headers)
        if r.status_code == 200:
            reports = json.loads(r.text).get('reports')
            for item in reports:
                if item.get('module') == module:
                    return not_empty_file(options['WORKSPACES'] + "/" + item.get('report'))
            
        #checking for each command of module
        url = options['REMOTE_API'] + "/api/routines?module=" + module
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

def is_force(options, filename):
    if options['FORCE'] != "False":
        return True

    if not_empty_file(filename):
        print_info(
            "Command is already done. use '-f' options to force rerun the command")
        return True
    return False

# checking if command was done or not? and return a json result
def checking_done(options, cmd=None, module=None, get_json=False):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/activities"

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

#force to update activities to all Done
def force_done(options, module):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/activities?module=" + module
    r = requests.put(url, verify=False, headers=headers)
    

def looping(options, cmd=None, module=None, times=5):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/activities"
    while times != 0:
        done = checking_done(options, cmd, module, url)
        if done:
            return
        times -= 1


def update_activities(options, data):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/activities"
    non_json_headers = headers
    non_json_headers['Content-Type'] = "text/html; charset=utf-8"
    # data = quote(str(data))
    # r = requests.patch(url, proxies=PROXY, verify=False,
    #                    headers=non_json_headers, data={'data': data})
    r = requests.patch(
    	url, verify=False, headers=non_json_headers, data={'data': data})


#just for conclusion
def save_all_cmd(options, logfile, module=None):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/activities"

    if module:
        url += '?module=' + module

    r = requests.get(url, verify=False, headers=headers)
    with open(logfile, 'w+') as l:
        l.write(r.text)
    # commands = json.loads(r.text)['commands']


def get_jwt(options):
    url = options['REMOTE_API'] + "/api/auth"
    username = options['USERNAME']
    password = options['PASSWORD']
    #set workspaces
    data = {'username': username, 'password': password}
    r = requests.post(url, verify=False, headers=headers, json=data)

    if r.status_code == 200:
        if json.loads(r.text).get('access_token'):
            print_good("Authentication success")
            token = "Bearer " + json.loads(r.text).get('access_token')
            return token
    return False


def set_config(options):
    url = options['REMOTE_API'] + "/api/config"
    #set workspaces
    data = {'options': options}
    r = requests.post(url, verify=False, headers=headers, json=data)
    return r


def just_shutdown_flask(options):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/shutdown"
    requests.post(url)
