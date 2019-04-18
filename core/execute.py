import sys, os, json
import subprocess, requests
# import utils

headers = {"User-Agent": "Osmedeus/v1.2", "Accept": "*/*",
           "Content-type": "application/json", "Connection": "close"}

def run1(command):
    os.system(command)

def run(command):
    stdout = ''
    try:
        process = subprocess.Popen(
            command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

        # Poll process for new output until finished
        while True:
            nextline = process.stdout.readline().decode('utf-8')
            # store output to log file
            if nextline == '' and process.poll() is not None:
                break
            print(nextline, end='')
            stdout += nextline
            sys.stdout.flush()

        exitCode = process.returncode

        if (exitCode == 0):
            return stdout
        else:
            print('Something went wrong with the command below: ')
            print(command)
            return None
    except:
        print('Something went wrong with the command below: ')
        print(command)
        return None
#get all commaands by module
def get_commands(options, module):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/routines?module=" + module

    r = requests.get(url, headers=headers)
    if r.status_code == 200:
        return json.loads(r.text)

    return None

def send_cmd(options, cmd, output_path='', std_path='', module='', nolog=False):
    headers['Authorization'] = options['JWT']
    # url = options['REMOTE_API'] + "/api/activities"
    json_cmd = {}
    json_cmd['cmd'] = cmd
    json_cmd['output_path'] = output_path
    json_cmd['std_path'] = std_path
    json_cmd['module'] = module
    #don't push this to activities log
    json_cmd['nolog'] = str(nolog)
    
    send_JSON(options, json_cmd)


#leave token blank for now
def send_JSON(options, json_body, token=''):
    headers['Authorization'] = options['JWT']
    url = options['REMOTE_API'] + "/api/cmd"
    #ignore the timeout
    try:
        r = requests.post(url, headers=headers, json=json_body, timeout=0.1)
    except:
        pass


