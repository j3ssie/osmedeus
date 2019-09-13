import os
import sys
import glob
import json
import requests
import time
import ipaddress
import socket
import shutil
import random
import hashlib
import base64
import re
import copy
from ast import literal_eval
from pathlib import Path
from itertools import chain
import inspect
import uuid
from bs4 import BeautifulSoup
import xml.etree.ElementTree as ET
from pprint import pprint
from configparser import ConfigParser
import tldextract
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

# Define some path
current_path = os.path.dirname(os.path.realpath(__file__))
ROOT_PATH = os.path.dirname(os.path.dirname(current_path))
DEAFULT_CONFIG_PATH = str(Path.home().joinpath(
    '.osmedeus/server.conf'))

TEMPLATE_SERVER_CONFIG = os.path.join(current_path, 'template-server.conf')
TEMPLATE_CLIENT_CONFIG = os.path.join(current_path, 'template-client.conf')

# send request through Burp proxy for debug purpose
PROXY = {
    'http': 'http://127.0.0.1:8081',
    'https': 'http://127.0.0.1:8081'
}


def print_load(text):
    print(f'{GR}' + '-'*70)
    print(f'{GR}{" ":10}{GR}^{B}O{GR}^ {G}{text} {GR}^{B}O{GR}^')
    print(f'{GR}' + '-'*70)


def print_block(text, tag='LOAD'):
    print(f'{GR}' + '-'*70)
    print(f'{GR}[{B}{tag}{GR}] {G}{text}')
    print(f'{GR}' + '-'*70)


def print_banner(text):
    print_block(text, tag='RUN')


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


def print_line():
    print(GR + '-' * 70)


def get_perf_time():
    return time.perf_counter()


def print_elapsed(options):
    current_module = options.get('CURRENT_MODULE', '')
    elapsed = time.perf_counter() - options.get('start_time')
    print(f"{GR}[{G}ESTIMATED{GR}] {C}{current_module}{GR} module executed in {C}{elapsed:0.2f}{GR} seconds.")


def print_debug(text, options=None):
    if not options:
        return
    if options.get('DEBUG'):
        print(G + "#" * 20 + GR)
        print(text)
        print("#" * 20)


def check_output(output):
    if output is None:
        return
    if os.path.isfile(output):
        if str(output) != '' and str(output) != "None":
            print('{1}--==[ Check the output: {2}{0}{1}'.format(output, G, P))
    elif os.path.isdir(output) and not_empty_file(output):
        print('{1}--==[ Check the output: {2}{0}{1}'.format(output, G, P))


def random_sleep(min=2, max=5, fixed=None):
    if fixed:
        time.sleep(fixed)
    else:
        time.sleep(random.randint(min, max))


# checking connection on port
def connection_check(target, port):
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        result = sock.connect_ex((target, int(port)))
        if result == 0:
            return True
        else:
            return False
    except Exception:
        return False


'''
### String utils
'''


def loop_grep(currents, source):
    for current in currents:
        for i in range(len(current)):
            if current[:i+1].lower().strip() == source:
                return True
    return False


def regex_strip(regex, string_in):
    real_regex = re.compile("{0}".format(regex))
    result = re.sub(real_regex, "", string_in)
    return result


def get_classes(class_name):
    return inspect.getmembers(sys.modules[class_name], inspect.isclass)


def get_methods(class_object, prefix=None):
    methods = [attr for attr in dir(class_object) if inspect.ismethod(
        getattr(class_object, attr))]
    if not prefix:
        return methods
    return [m for m in methods if m.startswith(prefix)]


def any_in(string_in, lists):
    return any(x in string_in for x in lists)


def upper_dict_keys(options):
    final_options = {}
    for key in options.keys():
        final_options[key.upper()] = options.get(key)
    return final_options


def lower_dict_keys(options):
    final_options = {}
    for key in options.keys():
        final_options[key.lower()] = options.get(key)
    return final_options


# clean some dangerous string before pass it to eval function
def safe_eval(base, inject_string, max_length=40):
    if len(inject_string) > max_length:
        return False
    if '.' in inject_string or ';' in inject_string:
        return False
    if '(' in inject_string or '}' in inject_string:
        return False
    if '%' in inject_string or '"' in inject_string:
        return False
    try:
        inject_string.encode('ascii')
    except:
        return False
    return base.format(inject_string)


# Yield successive n-sized chunks from l
def chunks(l, n):
    for i in range(0, len(l), n):
        yield l[i:i + n]


# just make sure there is a path
def clean_path(fpath):
    return os.path.normpath(fpath)


def get_tld(string_in):
    try:
        result = tldextract.extract(string_in)
        return result.domain
    except:
        return string_in


def get_uuid():
    return uuid.uuid4().hex


def strip_slash(string_in):
    if '/' in string_in:
        string_in = string_in.replace('/', '_')
    return string_in


# gen checksum field
def gen_checksum(string_in):
    if type(string_in) != str:
        string_in = str(string_in)
    checksum = hashlib.sha256("{0}".format(string_in).encode()).hexdigest()
    return checksum


def gen_ts():
    ts = int(time.time())
    return ts


# @TODO check XXE here
def is_xml(string_in, strict=True):
    if strict:
        # kind of prevent XXE
        if '<!ENTITY' in string_in or ' SYSTEM ' in string_in:
            return False
    try:
        tree = ET.fromstring(string_in)
        return True
    except:
        return False
    return False


def isBase64(sb):
    try:
        if type(sb) == str:
            sb_bytes = bytes(sb, 'ascii')
        elif type(sb) == bytes:
            sb_bytes = sb
        else:
            raise ValueError("Argument must be string or bytes")
        return base64.b64encode(base64.b64decode(sb_bytes)) == sb_bytes
    except Exception:
        return False


def is_json(string_in):
    try:
        json_object = json.loads(string_in)
    except:
        try:
            if type(literal_eval(string_in)) == dict:
                return True
        except:
            return False
    return True


def isURL(string_in):
    if 'http' in string_in:
        return True
    return False


def dict2json(dict_in):
    if type(dict_in) == dict:
        return json.dumps(dict_in)
    else:
        return dict_in


# check if string is IP or not
def valid_ip(string_in):
    try:
        ipaddress.ip_interface(str(string_in).strip())
        return True
    except:
        return False


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


# just get main domain
def get_domain(string_in):
    parsed = urllib.parse.urlparse(string_in)
    domain = parsed.netloc if parsed.netloc else parsed.path
    return domain


# get workspace name from options or direct string to prevent LFI
def get_workspace(options=None, workspace=None):
    if not workspace:
        raw_workspace = options.get('WORKSPACE')
    else:
        raw_workspace = workspace
    ws_name = os.path.basename(os.path.normpath(raw_workspace))
    return ws_name


def set_value(default, value):
    if value and value != '' and value != 'None':
        return value
    else:
        return default


# parsing xml string
def just_parse_xml(xml_string, get_dict=False):
    if is_xml(xml_string):
        root = ET.fromstring(xml_string)
        return root
    else:
        return False


# duplicate dict without reference
def just_copy(dict_in):
    return copy.deepcopy(dict_in)


def absolute_path(raw_path):
    return str(Path(raw_path).expanduser())


def file_copy(src, dest):
    shutil.copyfile(absolute_path(src), absolute_path(dest))


def just_chain(g1, g2):
    return chain(g1, g2)


def get_json(text):
    if is_json(text):
        if type(text) == dict:
            return text
        elif type(text) == str:
            try:
                return json.loads(text)
            except:
                return literal_eval(text)
    elif type(text) == dict:
        return text
    return False


def get_enviroment(env_name):
    return str(os.getenv(env_name))


def get_query(url):
    return urllib.parse.urlparse(url).query


def just_url_encode(string_in):
    return urllib.parse.quote(string_in)


def just_url_decode(string_in):
    return urllib.parse.unquote(string_in)


def just_b64_encode(string_in, encode_dict=False):
    if string_in:
        if encode_dict:
            if type(string_in) == dict:
                string_in = json.dumps(string_in)
                return base64.b64encode(string_in.encode()).decode()
            else:
                return base64.b64encode(string_in.encode()).decode()
        else:
            if type(string_in) != str:
                string_in = str(string_in)
            return base64.b64encode(string_in.encode()).decode()
    else:
        return string_in


def just_b64_decode(string_in, get_dict=False):
    if not string_in:
        return ''
    elif isBase64(string_in):
        if get_dict:
            string_out = base64.b64decode(string_in.encode()).decode()
            if type(string_out) == dict:
                return string_out
            elif is_json(string_out):
                return json.loads(string_out)
            elif type(literal_eval(string_out.strip('"'))) == dict:
                return literal_eval(string_out.strip('"'))
            else:
                return string_out
        return base64.b64decode(string_in.encode()).decode()
    else:
        return string_in


# just beatiful soup the xml or html
def soup(content, content_type='xml'):
    soup = BeautifulSoup(content, content_type)
    return soup


# join the request to
def url_join(url_dict, full_url=False):
    raw = urllib.parse.urlparse('/')
    # ('scheme', 'netloc', 'path', 'params', 'query', 'fragment')
    if url_dict.get('scheme'):
        raw = raw._replace(scheme=url_dict.get('scheme'))
    if url_dict.get('netloc'):
        raw = raw._replace(netloc=url_dict.get('netloc'))
    if url_dict.get('path'):
        raw = raw._replace(path=url_dict.get('path'))
    if url_dict.get('query'):
        raw = raw._replace(query=url_dict.get('query'))
    if url_dict.get('fragment'):
        raw = raw._replace(fragment=url_dict.get('fragment'))

    final = raw.geturl()
    if full_url:
        return final

    host = url_parse(final).netloc.split(':')[0]
    url = final.split(url_dict.get('netloc'))[1]
    return host, url


# parsing url
def url_parse(string_in, get_dict=False):
    parsed = urllib.parse.urlparse(string_in)
    if not get_dict:
        return parsed
    else:
        parsed_dict = {
            'scheme': parsed.scheme,
            'netloc': parsed.netloc,
            'path': parsed.path,
            'params': parsed.params,
            'query': parsed.query,
            'fragment': parsed.fragment,
        }
        return parsed_dict


# resolve list of commands
def resolve_commands(options, commands):
    results = []
    for raw_command in commands:
        command = just_copy(raw_command)
        for key, value in raw_command.items():
            command[key] = replace_argument(options, str(value))
        results.append(command)
    return results


# resolve list of commands
def resolve_command(options, raw_command):
    command = just_copy(raw_command)
    for key, value in raw_command.items():
        command[key] = replace_argument(options, str(value))
    return command


def check_required(command):
    if not command.get('requirement') or command.get('requirement') == '':
        return True
    if not not_empty_file(command.get('requirement')):
        print_bad("Requirement not found: {0}".format(command.get('requirement')))
        return False

    if not_empty_file(command.get('cleaned_output')):
        print_info("Post routine already done")
        return False

    return True


# replace argument in the command
def replace_argument(options, cmd):
    for key, value in options.items():
        if key in cmd:
            cmd = cmd.replace('$' + str(key), str(value))
    return cmd


'''
### End of string utils
'''


'''
File utils
'''


def get_parent(path):
    return os.path.dirname(path)


def join_path(parent, child):
    parent = os.path.normpath(parent)
    child = os.path.normpath(child)
    return os.path.join(parent, child.strip('/'))


def make_directory(directory, verbose=False):
    if directory and not os.path.exists(directory):
        if verbose:
            print_good('Make new directory: {0}'.format(directory))
        os.makedirs(directory)


def list_all(folder, ext='xml'):
    folder = os.path.normpath(folder)
    if os.path.isdir(folder):
        return glob.iglob(folder + '/**/*.{0}'.format(ext), recursive=True)
    return None


def just_write(filename, data, is_json=False, uniq=False, verbose=False):
    if not filename or not data:
        return False
    filename = os.path.normpath(filename)
    try:
        if verbose:
            print_good("Writing {0}".format(filename))
        if is_json:
            with open(filename, 'w+') as f:
                json.dump(data, f)
        else:
            with open(filename, 'w+') as f:
                f.write(data)
        return filename
    except:
        print_bad("Writing fail: {0}".format(filename))
        return False


def just_append(filename, data, is_json=False, uniq=False, verbose=False):
    if not filename or not data:
        return False
    filename = os.path.normpath(filename)
    try:
        if verbose:
            print_good("Writing {0}".format(filename))
        if is_json:
            with open(filename, 'a+') as f:
                json.dump(data, f)
        else:
            with open(filename, 'a+') as f:
                f.write(data)
        return filename
    except:
        print_bad("Writing fail: {0}".format(filename))
        return False


# unique and clean up
def clean_up(filename):
    if not filename:
        return False

    filename = os.path.normpath(filename)
    if os.path.isfile(filename):
        with open(filename, 'r') as f:
            data = f.readlines()

    real_data = list(set(data))
    just_write(filename, "\n".join(real_data))
    return True


def unique_list(list_in):
    if type(list_in) != list:
        return False
    return list(set(list_in))


def just_append(filename, data, is_json=False):
    if not filename:
        return False
    filename = os.path.normpath(filename)
    try:
        # print_good("Writing {0}".format(filename))
        if is_json:
            with open(filename, 'a+') as f:
                json.dump(data, f)
        else:
            with open(filename, 'a+') as f:
                f.write(data)
        return filename
    except:
        print_bad("Writing fail: {0}".format(filename))
        return False


def just_read_config(config_file, raw=False):
    if not not_empty_file(config_file):
        return False

    config = ConfigParser()
    config.read(config_file)
    if raw:
        return config
    sections = config.sections()
    options = {'CONFIG_PATH': config_file}
    for sec in sections:
        for key in config[sec]:
            options[key.upper()] = config.get(sec, key)
        # return options
    return options


def not_empty_file(filepath):
    if not filepath:
        return False
    fpath = os.path.normpath(filepath)
    return os.path.isfile(fpath) and os.path.getsize(fpath) > 0


def isFile(filepath):
    if os.path.isfile(filepath):
        return not_empty_file(filepath)
    else:
        return False


def just_read(filename, get_json=False, get_list=False):
    if not filename:
        return False
    filename = os.path.normpath(filename)
    if os.path.isfile(filename):
        with open(filename, 'r') as f:
            data = f.read()
        if get_json and is_json(data):
            return json.loads(data)
        elif get_list:
            return data.splitlines()
        return data

    return False


def strip_blank_line(filename, output):
    content = just_read(filename, get_list=True)
    if not content:
        return False
    output = os.path.normpath(output)

    with open(output, 'w+') as file:
        for line in content:
            if line.strip():
                file.write(line + "\n")
    return output


def get_ws(target):
    if not target:
        return False
    if os.path.isfile(target):
        return strip_slash(os.path.basename(target))
    else:
        return strip_slash(target)


def get_output_path(commands):
    output_list = []
    for command in commands:
        if command.get('cleaned_output') and not_empty_file(command.get('cleaned_output')):
            output_list.append(command.get('cleaned_output'))
        elif command.get('output_path') and not_empty_file(command.get('output_path')):
            output_list.append(command.get('output_path'))
    return output_list


def join_files(file_paths, output, uniq=True):
    if not file_paths and not output:
        return False
    if not uniq:
        with open(output, 'w') as outfile:
            for fname in file_paths:
                with open(fname) as infile:
                    for line in infile:
                        outfile.write(line)
        return output

    with open(output, 'w') as outfile:
        seen = set()
        for fname in file_paths:
            with open(fname) as infile:
                for line in infile:
                    if line not in seen:
                        outfile.write(line)
                        seen.add(line)
        return output


def is_done(options, final_output):
    if options.get('FORCED'):
        return False

    if not final_output:
        return False

    if type(final_output) == list:
        for report in final_output:
            if not not_empty_file(report):
                return False
        return True
    else:
        if not_empty_file(final_output):
            return True


# -gobuster.txt
def list_files(folder, pattern, empty_checked=True):
    if os.path.isfile(folder):
        folder = os.path.dirname(folder)
    folder = os.path.normpath(folder)
    if pattern:
        if pattern.startswith('**'):
            files = glob.iglob(folder + '/{0}'.format(pattern))
        else:
            files = glob.iglob(folder + '/**{0}'.format(pattern))

    if not empty_checked:
        return files
    return [f for f in files if not_empty_file(f)]
