import os
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.core import utils

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


# Just a helper message
def list_module_():
    print(''' 
List module
===========
subdomain   - Scanning subdomain and subdomain takeover
portscan    - Screenshot and Scanning service for list of domain
screenshot  - Screenshot list of hosts
vuln        - Scanning version of services and checking vulnerable service
git         - Scanning for git repo
burp        - Scanning for burp state
dirb        - Do directory search on the target
ip          - IP discovery on the target

python3 osmedeus.py -m <module> [-i <input>|-I <input_file>] [-t workspace_name]
        ''')
    sys.exit(0)


# Custom help message
def custom_help_():
    utils.print_info(
        "Visit this page for complete usage: https://j3ssie.github.io/Osmedeus/")
    print('''{1}
{2}Basic Usage{1}
===========
python3 osmedeus.py -t <your_target>
python3 osmedeus.py -T <list_of_targets>
python3 osmedeus.py -m <module> [-i <input>|-I <input_file>] [-t workspace_name]
python3 osmedeus.py --report <mode> -t <workspace> [-m <module>]

{2}Advanced Usage{1}
==============
{0}[*] List all module{1}
python3 osmedeus.py -M

{0}[*] List all report mode{1}
python3 osmedeus.py --report help

{0}[*] Running with specific module{1}
python3 osmedeus.py -t <result_folder> -m <module_name> -i <your_target>

{0}[*] Example command{1}
python3 osmedeus.py -m subdomain -t example.com
python3 osmedeus.py -t example.com --slow "subdomain"
python3 osmedeus.py -t sample2 -m vuln -i hosts.txt
python3 osmedeus.py -t sample2 -m dirb -i /tmp/list_of_hosts.txt

{2}Remote Options{1}
==============
--remote REMOTE       Remote address for API, (default: https://127.0.0.1:5000)
--auth AUTH           Specify authentication e.g: --auth="username:password"
                      See your config file for more detail (default: {2}core/config.conf{1})

--client              just run client stuff in case you already ran the Django server before

{2}More options{1}
==============
--update              Update lastest from git

-c CONFIG, --config CONFIG    
                      Specify config file (default: {2}core/config.conf{1})

-w WORKSPACE, --workspace WORKSPACE 
                      Custom workspace folder

-f, --force           force to run the module again if output exists
-s, --slow  "all"
                      All module running as slow mode         
-s, --slow  "subdomain"
                      Only running slow mode in subdomain module      

--debug               Just for debug purpose
            '''.format(G, GR, B))
    sys.exit(0)


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
