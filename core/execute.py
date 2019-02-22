import sys, os
import subprocess, requests
# import utils


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

def run_as_background(command):
    process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    return process


def send_cmd(cmd, output_path='', std_path='', module=''):
    json_cmd = {}
    json_cmd['cmd'] = cmd
    json_cmd['output_path'] = output_path
    json_cmd['std_path'] = std_path
    json_cmd['module'] = module
    
    send_JSON(json_cmd)



#leave token blank for now
def send_JSON(json_body, token=''):
    url = 'http://127.0.0.1:5000/cmd'
    headers = {"User-Agent": "Osmedeus/v1.0", "Accept": "*/*", "Content-type": "application/json", "Connection": "close"}
    # headers = {"User-Agent": "Osmedeus/v1.0", "Accept": "*/*", "Authorization": "Bearer " + token, "Content-type": "application/json", "Connection": "close"}
    #ignore the timeout
    try:
        r = requests.post(url, headers=headers, json=json_body, timeout=0.1)
    except:
        pass
    # return r


