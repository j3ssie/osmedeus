import os

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

def check_output(options, raw_output):
	output = replace_argument(options, raw_output)
	print('{1}--==[ Check the output: {2}{0}'.format(output, G, P))


def replace_argument(options, cmd):
	for key,value in options['env'].items():
		if key in cmd:
			cmd = cmd.replace('$' + str(key), str(value))	
	return cmd

def make_directory(directory):
	if not os.path.exists(directory):
		print_good('Make new workspace: {0}'.format(directory))
		os.makedirs(directory)

def not_empty_file(fpath):  
	return os.path.isfile(fpath) and os.path.getsize(fpath) > 0


#checking the workspace and plugin path
def initial_stuff(options):
	utils.make_directory(options['env']['WORKSPACE'])
	utils.make_directory(options['env']['WORKSPACE'] + '/subdomain')
	utils.make_directory(options['env']['WORKSPACE'] + '/portscan')
	utils.make_directory(options['env']['WORKSPACE'] + '/screenshot')
	utils.make_directory(options['env']['WORKSPACE'] + '/screenshot/all')

	utils.make_directory(options['env']['WORKSPACE'] + '/gitscan/')
	utils.make_directory(options['env']['WORKSPACE'] + '/bruteforce/')
	utils.make_directory(options['env']['WORKSPACE'] + '/directory/')
	utils.make_directory(options['env']['WORKSPACE'] + '/burpstate/')
	utils.make_directory(options['env']['WORKSPACE'] + '/vulnscan/')





