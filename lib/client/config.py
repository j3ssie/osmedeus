import os
import sys
import shutil
from pathlib import Path

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
from lib.core import execute
from lib.core import utils
from lib.client.banner import banner_
from lib.client.update import update_
from lib.client.helpers import list_module_
from lib.client.helpers import custom_help_


# Just print some help message
def banner(__version__, __author__):
    banner_(__version__, __author__)


def list_module():
    list_module_()


def custom_help():
    custom_help_()


def update():
    update_()


def _verify_target(target, target_list):
    if target_list:
        if utils.not_empty_file(target_list):
            real_target = target_list
        else:
            utils.print_bad("Input file not found: {0}".format(target_list))
            sys.exit(-1)
    else:
        real_target = target

    return real_target


def _config_file_handle(config_path, remote, credentials):
    # checking for config path
    if os.path.isfile(config_path):
        utils.print_info('Loading config file from: {0}'.format(config_path))
    else:
        utils.print_info('New config file created: {0}'.format(config_path))
        utils.file_copy(utils.TEMPLATE_CLIENT_CONFIG, config_path)

    configs = utils.just_read_config(config_path, raw=True)
    remote, credentials = _handle_remote(remote, credentials, configs)

    # write the config again
    configs.set('Server', 'remote_api', remote)
    configs.set('Server', 'username', credentials[0])
    configs.set('Server', 'password', credentials[1])
    with open(config_path, 'w+') as configfile:
        configs.write(configfile)

    return remote, credentials


# parsing remote api and credentials
def _handle_remote(remote, credentials, configs):
    # from config file
    config_remote = configs.get('Server', 'remote_api')
    config_username = configs.get('Server', 'username')
    config_password = configs.get('Server', 'password')

    # if defined from cli get it or get it from config file
    remote = config_remote if remote is None else remote

    if credentials and ':' in credentials:
        credentials = (credentials.split(':')[0].strip(), credentials.split(':')[1].strip())
    else:
        credentials = (config_username, config_password)
    return remote, credentials


# clean None options
def _clean_None(options):
    for key in utils.just_copy(options).keys():
        if options[key] is None:
            del options[key]
    return options


# parsing args
def parsing_config(args):
    # create default osmedeus client path
    osmedeus_home = str(Path.home().joinpath('.osmedeus'))
    utils.make_directory(osmedeus_home)

    # parsing remote server here
    remote = args.remote if args.remote else None
    credentials = args.auth if args.auth else None

    # reading old config file or create new one
    if args.config_path:
        config_path = args.config_path
    else:
        config_path = str(Path.home().joinpath('.osmedeus/client.conf'))
    remote, credentials = _config_file_handle(
        config_path, remote, credentials)

    # remote, credentials = _handle_remote(remote, credentials, configs)

    # folder name to store all the results
    workspace = args.workspace if args.workspace else None

    # Target stuff
    target = args.target if args.target else None
    target_list = args.targetlist if args.targetlist else None
    target = _verify_target(target, target_list)

    # get direct input as single or a file
    direct_input = args.input if args.input else None
    direct_input_list = args.inputlist if args.inputlist else None
    direct_input = _verify_target(direct_input, direct_input_list)

    # parsing speed config
    if args.slow and args.slow.lower() == 'all':
        speed = "quick|-;;slow|*"  # all slow
    elif args.slow:
        # quick all but some slow
        speed = "quick|*;;slow|{0}".format(args.slow.strip())
    else:
        speed = "quick|*;;slow|-"  # all quick

    # parsing modules
    modules = args.modules if args.modules else None
    exclude = args.exclude if args.exclude else ''

    localhost = args.localhost if args.localhost else False
    report = args.report if args.report else None
    # turn on default
    slack = False if args.noslack else True
    monitor = False if args.nomonitor else True

    if modules:
        if direct_input_list:
            mode = 'direct_list'
        else:
            mode = 'direct'
    elif report:
        mode = 'report'
    else:
        mode = 'general'
    # modules = _handle_speed(raw_modules, speed, mode)

    debug = str(args.debug)
    forced = str(args.forced)

    # select one
    if mode == 'general':
        real_target = utils.set_value(direct_input, target)
    else:
        real_target = utils.set_value(target, direct_input)

    options = {
        'start_ts': utils.gen_ts(),
        'raw_target': real_target,
        'target_list': target_list,
        'mode': mode,
        'slack': slack,
        'speed': speed,
        'workspace': workspace,
        'modules': modules,
        'exclude': exclude,
        'forced': forced,
        'debug': debug,
        'remote_api': remote.strip('/'),
        'credentials': credentials,
        'localhost': localhost,
        'report': report,
        'monitor': monitor,
    }
    # clean None options before send submit request
    options = _clean_None(options)
    return options
