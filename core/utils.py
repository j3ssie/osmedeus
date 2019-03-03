import os, json, requests, time
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
		print_good('Make new workspace: {0}'.format(directory))
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
    if os.path.isfile(filename):
        with open(filename, 'r') as f:
            main_json = json.load(f)
        return main_json

    return None

###
# checking if command was done or not? and return a json result
def checking_done(cmd=None, module=None, get_json=False, url='http://127.0.0.1:5000/activities'):
    if cmd:
        r = requests.post(url, headers=headers, json={'cmd' : cmd})
    if module:
        r = requests.post(url + "?module=" + module, headers=headers, json={})

    commands = json.loads(r.text)
    for cmd in commands['commands']:
        if cmd['status'] != 'Done':

            return False if not get_json else commands

    return True if not get_json else commands


def looping(cmd=None, module=None, times=5, url='http://127.0.0.1:5000/activities'):
    while times != 0:
        done = checking_done(cmd, module, url)
        if done:
            return

        times -= 1

#just for conclusion
def save_all_cmd(logfile, url='http://127.0.0.1:5000/activities'):
    r = requests.get(url)
    with open(logfile, 'w+') as l:
        l.write(r.text)
    # commands = json.loads(r.text)['commands']


def set_config(options, url='http://127.0.0.1:5000/config'):
    #set workspaces
    r = requests.post(url, headers=headers, json={'workspaces' : options['WORKSPACES']})

    return r

def just_shutdown_flask(url='http://127.0.0.1:5000/shutdown'):
    requests.post(url)


