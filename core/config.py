import os, sys, socket, time
import shutil, string, random
import urllib.parse


from core import execute
from core import utils
from pprint import pprint
from configparser import ConfigParser, ExtendedInterpolation
import string

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

__author__ = '@j3ssiejjj'
__version__ = '1.3'

def banner():
    print(r"""{1}

                       `@@`
                      @@@@@@
                    .@@`  `@@.
                    :@      @:
                    :@  {5}:@{1}  @:                       
                    :@  {5}:@{1}  @:                       
                    :@      @:                             
                    `@@.  .@@`
                      @@@@@@
                        @@
                     {0}@{1}  {1}@@  {0}@{1}               
                    {0}+@@{1} {1}@@ {0}@@+{1}                    
                 {5}@@:@#@,{1}{1}@@,{5}@#@:@@{1}           
                ;@+@@`#@@@@#`@@+@;
                @+ #@@@@@@@@@@# +@
               @@  @+`@@@@@@`+@  @@
               @.  @   ;@@;   @  .@
              {0}#@{1}  {0}'@{1}          {0}@;{1}  {0}@#{1}

                     
             Osmedeus v{5}{6}{1} by {2}{7}{1}

                    ¯\_(ツ)_/¯
        """.format(C, G, P, R, B, GR, __version__, __author__))


def update():
    execute.run1(
        'git fetch --all && git reset --hard origin/master && ./install.sh')
    sys.exit(0)


def list_module():
    print(''' 
List module
===========
subdomain   - Scanning subdomain and subdomain takerover
portscan    - Screenshot and Scanning service for list of domain
brute       - Do brute force on service of target
vuln        - Scanning version of services and checking vulnerable service
git         - Scanning for git repo
burp        - Scanning for burp state
dirb        - Do directory search on the target
ip          - IP discovery on the target

        ''')
    sys.exit(0)



def custom_help():
    utils.print_info("Visit this page for complete usage: https://github.com/j3ssie/Osmedeus/wiki")
    print('''{1}
{2}Basic Usage{1}
===========
python3 osmedeus.py -t <your_target>
python3 osmedeus.py -T <list_of_targets>

{2}Advanced Usage{1}
==============
{0}[*] List all module{1}
python3 osmedeus.py -M 

{0}[*] Running with specific module{1}
python3 osmedeus.py -t <result_folder> -m <module_name> -i <your_target>

{0}[*] Example command{1}
python3 osmedeus.py -t sample2 -m vuln -i hosts.txt
python3 osmedeus.py -t sample2 -m dirb -i /tmp/list_of_hosts.txt

{2}Remote Options{1}
==============
--remote REMOTE       Remote address for API, (default: https://127.0.0.1:5000)
--auth AUTH           Specify auth tication e.g: --auth="username:password"
                      See your config file for more detail (default: {2}core/config.conf{1})

--client              just run client stuff in case you ran the flask server before

{2}More options{1}
==============
--update              Update lastest from git

-c CONFIG, --config CONFIG    
                      Specify config file (default: {2}core/config.conf{1})

-w WORKSPACE, --workspace WORKSPACE 
                      Custom workspace folder

-f, --force           force to run the module again if output exists
-v, --verbose         show verbose output
-q, --quick           run this tool with quick routine
-s, --slow            run this tool with slow routine

--debug               Just for debug purpose
            '''.format(G, GR, B))
    sys.exit(0)

def proxy_parsing(options):
    # return if proxy config file found
    if options['PROXY_FILE'] != "None":
        proxy_file = options['PROXY_FILE']
        utils.print_info("Detected proxychains file: {0}".format(proxy_file))
        return 
    elif options['PROXY'] != "None":
        proxy_file = options['CWD'] + '/core/proxychains.conf'
        utils.print_info("Detected proxychains file: {0}".format(proxy_file))


    if options['PROXY'] != "None":
        proxy_parsed = urllib.parse.urlsplit(options['PROXY'])

        scheme = proxy_parsed.scheme
        host = proxy_parsed.netloc.split(':')[0]
        port = proxy_parsed.netloc.split(':')[1]

        proxy_element = "\n" + scheme + " " + host + " " + port

        raw_data = utils.just_read(proxy_file).splitlines()
        for i in range(len(raw_data)):
            if '[ProxyList]' in raw_data[i]:
                init_part = raw_data[:i]
                proxy_part = raw_data[i:]

        # check if this proxy is exist or not
        check_duplicate = False 
        for item in proxy_part:
            if proxy_element.strip() in item.strip():
                check_duplicate = True
        
        if not check_duplicate: 
            proxy_part.append(proxy_element)
        
        real_proxy_data = "\n".join(init_part + proxy_part)
        utils.just_write(proxy_file, real_proxy_data)
    
    if options['PROXY'] != "None" or options['PROXY_FILE'] != "None":
        if not shutil.which(options['PROXY_CMD'].split(' ')[0]):
            utils.print_bad("Look like proxy mode doesn't support your OS")
            sys.exit(0)
        else:
            #simple check for proxy is good
            utils.print_info("Testing proxy with simple curl command")
            if execute.run(options['PROXY_CMD'] + " curl -s ipinfo.io/ip") == execute.run("curl -s ipinfo.io/ip"):
                utils.print_bad("Look like your proxy not work properly")
                sys.exit(0)


def parsing_config(config_path, args):
    options = {}

    ##some default path
    github_api_key = str(os.getenv("GITROB_ACCESS_TOKEN"))
    cwd = str(os.getcwd())

    #just hardcode if gopath not loaded
    go_path = cwd + "/plugins/go"
    # go_path = str(os.getenv("GOPATH")) + "/bin"
    # if "None" in go_path:
    #     go_path = cwd + "/plugins/go"

    bot_token = str(os.getenv("SLACK_BOT_TOKEN"))
    log_channel = str(os.getenv("LOG_CHANNEL"))
    status_channel = str(os.getenv("STATUS_CHANNEL"))
    report_channel = str(os.getenv("REPORT_CHANNEL"))
    stds_channel = str(os.getenv("STDS_CHANNEL"))
    verbose_report_channel = str(os.getenv("VERBOSE_REPORT_CHANNEL"))


    if os.path.isfile(config_path):
        utils.print_info('Config file detected: {0}'.format(config_path))
        #config to logging some output
        config = ConfigParser(interpolation=ExtendedInterpolation())
        config.read(config_path)
    else:
        utils.print_info('New config file created: {0}'.format(config_path))
        shutil.copyfile(cwd + '/template-config.conf', config_path)

        config = ConfigParser(interpolation=ExtendedInterpolation())
        config.read(config_path)

    workspace = cwd + "/workspaces/"
    config.set('Enviroments', 'cwd', cwd)
    config.set('Enviroments', 'go_path', go_path)
    config.set('Enviroments', 'github_api_key', github_api_key)
    config.set('Enviroments', 'workspaces', str(workspace))

    if args.debug:
        config.set('Slack', 'bot_token', 'bot_token')
        config.set('Slack', 'log_channel', 'log_channel')
        config.set('Slack', 'status_channel', 'status_channel')
        config.set('Slack', 'report_channel', 'report_channel')
        config.set('Slack', 'stds_channel', 'stds_channel')
        config.set('Slack', 'verbose_report_channel', 'verbose_report_channel')
    else:
        config.set('Slack', 'bot_token', bot_token)
        config.set('Slack', 'log_channel', log_channel)
        config.set('Slack', 'status_channel', status_channel)
        config.set('Slack', 'report_channel', report_channel)
        config.set('Slack', 'stds_channel', stds_channel)
        config.set('Slack', 'verbose_report_channel', verbose_report_channel)

    ##config of the tool
    if args.slow:
        speed = "slow"
    else:
        speed = "quick"

    module = str(args.module)
    debug = str(args.debug)
    force = str(args.force)

    config.set('Mode', 'speed', speed)
    config.set('Mode', 'module', module)
    config.set('Mode', 'debug', debug)
    config.set('Mode', 'force', force)

    ##target stuff
    #parsing agument
    git_target = args.git if args.git else None
    burpstate_target = args.burp if args.burp else None
    target_list = args.targetlist if args.targetlist else None
    company = args.company if args.company else None
    direct_input = args.input if args.input else None

    if args.target:
        target = args.target
        output = args.output if args.output else args.target
        company = args.company if args.company else args.target

        strip_target = target.replace('https://', '').replace('http://', '')
        if '/' in strip_target:
            strip_target = strip_target.split('/')[0]

        if args.workspace:
            if args.workspace[-1] == '/':
                workspace = args.workspace + options['env']['STRIP_TARGET']
            else:
                workspace = args.workspace + '/' + options['env']['STRIP_TARGET']
        else:
            workspace += strip_target

        if not direct_input:
            try:
                ip = socket.gethostbyname(strip_target)
            except:
                ip = "None"
                utils.print_bad("Something wrong to connect to {0}".format(target))
        else:
            ip = None

    #getting proxy from args
    proxy = args.proxy if args.proxy else None
    proxy_file = args.proxy_file if args.proxy_file else None

    config.set('Proxy', 'proxy', str(proxy))
    config.set('Proxy', 'proxy_file', str(proxy_file))

    if config['Proxy']['proxy_cmd'] == 'None':
        #only works for Kali proxychains, change it if you on other OS
        proxy_cmd = "proxychains -f {0}".format(proxy_file)
        config.set('Proxy', 'proxy_cmd', str(proxy_cmd))


    config.set('Target', 'input', str(direct_input))
    config.set('Target', 'git_target', str(git_target))
    config.set('Target', 'burpstate_target', str(burpstate_target))
    config.set('Target', 'target_list', str(target_list))
    config.set('Target', 'output', str(output))
    config.set('Target', 'target', str(target))
    config.set('Target', 'strip_target', str(strip_target))
    config.set('Target', 'company', str(company))
    config.set('Target', 'ip', str(ip))

    config.set('Enviroments', 'workspace', str(workspace))

    #create workspace folder for the target
    utils.make_directory(workspace)

    #set the remote API
    if args.remote:
        remote_api = args.remote
        config.set('Server', 'remote_api', remote_api)
    
    #set credentials as you define from agurments
    if args.auth:
        #user:pass
        creds = args.auth.strip().split(":")
        username = creds[0]
        password = creds[1]
        
        config.set('Server', 'username', username)
        config.set('Server', 'password', password)
    else:
        #set random password if default password detect
        if config['Server']['password'] == 'super_secret':
            new_pass = ''.join(random.choice(string.ascii_lowercase)
                            for i in range(6)).upper()
            config.set('Server', 'password', new_pass)

    #save the config
    with open(config_path, 'w') as configfile:
        config.write(configfile)

    config = ConfigParser(interpolation=ExtendedInterpolation())
    config.read(config_path)
    sections = config.sections()

    for sec in sections:
        for key in config[sec]:
            options[key.upper()] = config.get(sec, key)

    ######
    #parsing proxy stuff
    proxy_parsing(options)

    return options

