import os
import sys
import time
import subprocess
import itertools
sys.path.append(os.path.dirname(os.path.realpath(__file__)))

from core import utils


# just run command directly for simple purpose like update
def run1(command):
    os.system(command)


# run command in subprocess
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
            utils.print_line()
            print('Something went wrong with the command below: ')
            print(command)
            utils.print_line()
            return None
    except:
        utils.print_line()
        utils.print_bad('Something went wrong with the command below: ')
        print(command)
        utils.print_line()
        return None


def run_single(command):
    forced = command.get('forced', False)
    if not forced:
        if utils.not_empty_file(command.get('output_path')):
            return True
    std_out = run(command.get('cmd'))
    # store std and output
    if command.get('std_path') != '':
        utils.just_write(command.get('std_path'), std_out)
    if command.get('output_path') != '':
        utils.check_output(command.get('output_path'))
    return True


def run_chunk(commands, chunk, delay=1):
    for part in utils.chunks(commands, int(chunk)):
        done = False
        for command in part:
            done = run_single(command)
        if done:
            time.sleep(0.5)
        else:
            time.sleep(int(delay))


def parse_resources(resources):
    if not resources or resources == '':
        return False
    # resources = 'l0|/tmp/sam.txt;;l1|/tmp/as'
    # l0 is a input and
    if ';;' in resources:
        raw_lists = resources.split(';;')
    else:
        raw_lists = [resources]

    _lists = []
    for item in raw_lists:
        if not item or item == '':
            continue
        if '|' in item:
            filename = ''.join(item.split('|')[1:])
            content = utils.just_read(filename, get_list=True)
            _lists.append({
                'filename': filename,
                'content': content,
                })

    # currently only support 2 list
    if len(_lists) == 1:
        final_lists = list(zip(_lists[0].get('content'), itertools.cycle([''])))
    if len(_lists) == 2:
        final_lists = list(zip(_lists[0].get('content'), itertools.cycle(_lists[1].get('content'),)))

    return final_lists


def really_replace(string_in, item):
    string_in = string_in.replace('[[0]]', item[0])
    string_in = string_in.replace('[[1]]', item[1])
    return string_in


# gen commands base on resource list
def get_chunk_commands(command):
    lists = parse_resources(command.get('resources'))
    # print(lists)
    if not lists:
        return False
    # content1 = utils.just_read(resources, get_list=True)

    commands = []

    for ele in lists:
        item = utils.just_copy(command)
        item['cmd'] = really_replace(command.get('cmd'), ele)
        item['output_path'] = really_replace(command.get('output_path'), ele)
        item['std_path'] = really_replace(command.get('std_path'), ele)
        # print(item)
        commands.append(utils.just_copy(item))

    # print(commands)

    return commands
