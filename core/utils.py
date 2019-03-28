import os, json, requests, time
from urllib.parse import quote, unquote

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
colors = [G,R,B,P,C,O,GR]

info = '{0}[*]{1} '.format(B,W)
ques =  '{0}[?]{1} '.format(C,W)
bad = '{0}[-]{1} '.format(R,W)
good = '{0}[+]{1} '.format(G,W)

headers = {"User-Agent": "Osmedeus/v1.0", "Accept": "*/*",
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
	print('{1}--==[ Check the output: {2}{0}{1}'.format(output, G, P))

#################

def replace_argument(options, cmd):
	# for key,value in options['env'].items():
	for key,value in options.items():
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

def just_waiting(module_name, seconds=30):
    while not checking_done(module=module_name):
        print_info('Waiting for {0} module'.format(module_name))
        time.sleep(seconds)


def reading_json(filename):
    try:
        if os.path.isfile(filename):
            with open(filename, 'r') as f:
                main_json = json.load(f)
            return main_json
    except:
        print_bad("Reading fail: {0}".format(filename))

    return None

###
# checking if command was done or not? and return a json result
def checking_done(cmd=None, module=None, get_json=False, url='http://127.0.0.1:5000/api/activities'):
    if cmd:
        r = requests.post(url, headers=headers, json={'cmd' : cmd})
    if module:
        r = requests.post(url + "?module=" + module, headers=headers, json={})

    commands = json.loads(r.text)
    for cmd in commands['commands']:
        if cmd['status'] != 'Done':

            return False if not get_json else commands

    return True if not get_json else commands


def looping(cmd=None, module=None, times=5, url='http://127.0.0.1:5000/api/activities'):
    while times != 0:
        done = checking_done(cmd, module, url)
        if done:
            return
        times -= 1


def update_activities(data, url='http://127.0.0.1:5000/api/activities'):
    data = quote(str(data))
    # r = requests.patch(url, headers=headers, data=data)
    r = requests.patch(url, headers={"User-Agent": "Osmedeus/v1.0"}, data={'data': data}, proxies=PROXY)

#just for conclusion
def save_all_cmd(logfile, module=None, url='http://127.0.0.1:5000/api/activities'):
    if module:
        url += '?module=' + module
    
    r = requests.get(url, headers=headers)
    with open(logfile, 'w+') as l:
        l.write(r.text)
    # commands = json.loads(r.text)['commands']


def set_config(options, url='http://127.0.0.1:5000/api/config'):
    #set workspaces
    data = {'options': options}
    r = requests.post(url, headers=headers, json=data)
    return r

def just_shutdown_flask(url='http://127.0.0.1:5000/api/shutdown'):
    requests.post(url)


