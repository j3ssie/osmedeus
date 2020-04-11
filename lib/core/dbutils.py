from api.models import *
from workflow import general, direct, direct_list
from core import utils


def parse_special_line(line):
    parts = line.split(';;')
    jsonl = {}
    for part in parts:
        key = part.split('|')[0].lower()
        value = ''.join(part.split('|')[1:])
        jsonl[key] = value
    return jsonl


def update_field(default, value, update_type='partial'):
    final_value = ''
    if update_type.lower() == 'full':
        final_value = value
    if default == 'N/A' or default == '' or default == 'None':
        final_value = value.strip(',')
    else:
        final_value = default.strip(',') + ',' + value.strip(',')

    if ',' in final_value:
        results = [x.strip() for x in final_value.split(',')]
        final_value = ",".join(list(set(results)))
    return final_value


def clean_up(record):
    model_fields = ['domain', 'ip_address', 'technologies', 'ports', 'workspace', 'paths', 'screenshot', 'note', 'checksum']
    keys = list(record.keys())
    for key in keys:
        if key not in model_fields:
            del record[key]
    return record


def parse_summary_field(instance, jsonl, update_type):
    # just something we don't want to update
    blacklist = ['domain']
    record = instance.as_json()
    checksum  = record.get('checksum')
    for key, value in jsonl.items():
        if key not in blacklist:
            record[key] = update_field(record[key], str(value), update_type)
    record = clean_up(record)
    updated = Summaries.objects.filter(checksum=checksum).update(**record)


def import_domain_summary(jsonl, workspace, update_type):
    # print(jsonl)
    domain = jsonl.get('domain', None)
    if domain is None:
        domain = jsonl.get('ip_address')
    instance, created = Summaries.objects.get_or_create(domain=domain, workspace=workspace)

    parse_summary_field(instance, jsonl, update_type)


# Summaries part
def parse_domains(line):
    if utils.is_json(line.strip()):
        jsonl = utils.get_json(line)
    elif ';;' in line.strip():
        jsonl = parse_special_line(line)
    else:
        jsonl = {'domain': line.strip()}
    return jsonl


# remove report part
def removeReport(speed):
    if speed.lower() in 'report':
        return True
    else:
        return False


def clean_input(raw_input, module='general'):
    if 'general' in module.lower():
        return utils.get_domain(raw_input)

    elif 'dir' in module.lower():
        return raw_input


def gen_default_config(config_path):
    config_path = utils.absolute_path(config_path)
    utils.file_copy(utils.TEMPLATE_SERVER_CONFIG, config_path)

    configs = utils.just_read_config(config_path, raw=True)

    workspaces = utils.join_path(utils.get_parent(
        utils.DEAFULT_CONFIG_PATH), 'workspaces')
    plugins_path = utils.join_path(utils.ROOT_PATH, 'plugins')
    go_path = utils.join_path(utils.ROOT_PATH, 'plugins/go')
    data_path = utils.join_path(utils.ROOT_PATH, 'data')
    alias_path = utils.join_path(utils.ROOT_PATH, 'lib/alias')

    # set some path
    configs.set('Enviroments', 'workspaces', workspaces)
    configs.set('Enviroments', 'plugins_path', plugins_path)
    configs.set('Enviroments', 'data_path', data_path)
    configs.set('Enviroments', 'alias_path', alias_path)
    configs.set('Enviroments', 'go_path', go_path)

    # set some tokens
    github_api_key = utils.get_enviroment("GITHUB_API_KEY")
    slack_bot_token = utils.get_enviroment("SLACK_BOT_TOKEN")
    log_channel = utils.get_enviroment("LOG_CHANNEL")
    status_channel = utils.get_enviroment("STATUS_CHANNEL")
    report_channel = utils.get_enviroment("REPORT_CHANNEL")
    stds_channel = utils.get_enviroment("STDS_CHANNEL")
    verbose_report_channel = utils.get_enviroment("VERBOSE_REPORT_CHANNEL")
    configs.set('Enviroments', 'github_api_key', github_api_key)
    configs.set('Slack', 'slack_bot_token', slack_bot_token)
    configs.set('Slack', 'log_channel', log_channel)
    configs.set('Slack', 'status_channel', status_channel)
    configs.set('Slack', 'report_channel', report_channel)
    configs.set('Slack', 'stds_channel', stds_channel)
    configs.set('Slack', 'verbose_report_channel', verbose_report_channel)

    telegram_bot_token = utils.get_enviroment("TELEGRAM_BOT_TOKEN")
    telegram_log_channel = utils.get_enviroment("TELEGRAM_LOG_CHANNEL")
    telegram_status_channel = utils.get_enviroment("TELEGRAM_STATUS_CHANNEL")
    telegram_report_channel = utils.get_enviroment("TELEGRAM_REPORT_CHANNEL")
    telegram_stds_channel = utils.get_enviroment("TELEGRAM_STDS_CHANNEL")
    telegram_verbose_report_channel = utils.get_enviroment("TELEGRAM_VERBOSE_REPORT_CHANNEL")
    configs.set('Telegram', 'telegram_bot_token', telegram_bot_token)
    configs.set('Telegram', 'telegram_log_channel', telegram_log_channel)
    configs.set('Telegram', 'telegram_status_channel', telegram_status_channel)
    configs.set('Telegram', 'telegram_report_channel', telegram_report_channel)
    configs.set('Telegram', 'telegram_stds_channel', telegram_stds_channel)
    configs.set('Telegram', 'telegram_verbose_report_channel', telegram_verbose_report_channel)

    # monitor mode
    backups = utils.join_path(utils.get_parent(
        utils.DEAFULT_CONFIG_PATH), 'backups')
    utils.make_directory(backups)

    monitors = utils.join_path(utils.get_parent(
        utils.DEAFULT_CONFIG_PATH), 'monitors')
    utils.make_directory(monitors)

    configs.set('Monitor', 'monitors', monitors)
    configs.set('Monitor', 'backups', backups)
    monitor_level = utils.get_enviroment("monitor_level", 'final')
    configs.set('Monitor', 'monitor_level', monitor_level)

    # monitor bot
    slack_monitor_token = utils.get_enviroment("SLACK_MONITOR_TOKEN")
    new_channel = utils.get_enviroment("NEW_CHANNEL")
    new_name = utils.get_enviroment("NEW_NAME")
    missing_channel = utils.get_enviroment("MISSING_CHANNEL")
    missing_name = utils.get_enviroment("MISSING_NAME")
    configs.set('Monitor', 'slack_monitor_token', slack_monitor_token)
    configs.set('Monitor', 'new_channel', new_channel)
    configs.set('Monitor', 'new_name', new_name)
    configs.set('Monitor', 'missing_channel', missing_channel)
    configs.set('Monitor', 'missing_name', missing_name)

    # write it again
    with open(config_path, 'w+') as configfile:
        configs.write(configfile)

    # read it again and return
    options = utils.just_read_config(config_path)
    return options


def load_default_config(config_file=None, forced_reload=False):
    if not config_file:
        config_file = '~/.osmedeus/server.conf'
    options = utils.just_read_config(config_file)

    # no config found generate one from default config
    if not options:
        options = gen_default_config(config_file)
    if forced_reload:
        options = gen_default_config(config_file)

    # looping and adding field to db
    for key, value in options.items():
        item = {
            'name': key,
            'value': value,
            'alias': key,
            'desc': key,
        }
        instance, created = Configurations.objects.get_or_create(
            name=key)
        Configurations.objects.filter(name=key).update(**item)
    return options


def get_stateless_options(config_file=None):
    if config_file:
        options = utils.just_read_config(config_file)
    else:
        raw_options = list(Configurations.objects.values_list('name', 'value'))
        options = {}
        for item in raw_options:
            options[item[0]] = item[1]
    return options


# get variable to replace in the command
def get_stateful_options(workspace):
    # finding workspace in db
    record = Workspaces.objects.filter(workspace=workspace)
    if not record.first():
        record = Workspaces.objects.filter(target=workspace)
    if not record.first():
        record = Workspaces.objects.filter(raw_target=workspace)

    if not record.first():
        return False
    stateless_options = get_stateless_options()

    # options = record.as_json()
    options = {**record.first().as_json(), **stateless_options}
    argument_options = {}

    # just upper all key
    for key in options.keys():
        argument_options[key.upper()] = options.get(key)

    return argument_options


# @TODO should be done dynamic later
def get_modules(mode='general'):
    general = [
        'SubdomainScanning',
        'Recon',
        'ScreenShot',
        'TakeOverScanning',
        'AssestFinding',
        'IPSpace',
        'CorsScan',
        'PortScan',
        'VulnScan'
    ]
    if 'general' in mode.lower():
        return ','.join(general)


# really parse command from classes
def really_commands(mode):
    modules = utils.get_classes('workflow.{0}'.format(mode))
    for module in modules:
        # get RCE if you can edit general file in workflow folder :)
        module_name = module[0].strip()
        module_object = eval('{0}.{1}'.format(mode, module_name))
        # parsing commands
        try:
            routines = module_object.commands
        except:
            continue
        for routine, commands in routines.items():
            for command in commands:
                item = command
                item['mode'] = mode
                item['speed'] = routine
                item['module'] = module_name
                item['alias'] = module_name + "__" + routine.lower() + "__" + \
                    str(item.get('banner')).lower()
                Commands.objects.create(**item)

        reports = module_object.reports
        parse_report(reports, module_name, mode)


def internal_parse_commands(override=True):
    if override:
        Commands.objects.all().delete()
        ReportsSkeleton.objects.all().delete()
    really_commands('general')
    really_commands('direct')
    really_commands('direct_list')


def parse_report(reports, module, mode):
    if type(reports) == str:
        item = {
            'report_path': reports,
            'report_type': 'bash',
            'module': module,
            'mode': mode,
        }
        ReportsSkeleton.objects.create(**item)
    elif type(reports) == list:
        for report in reports:
            item = {
                'report_path': report.get('path'),
                'report_type': report.get('type', 'bash'),
                'note': report.get('note', ''),
                'module': module,
                'mode': mode,
            }
            ReportsSkeleton.objects.create(**item)


# parsing skeleton commands
def parse_commands(command_path):
    if not utils.not_empty_file(command_path):
        return False

    content = utils.just_read(command_path, get_json=True)
    if not content:
        return False

    modules = content.keys()
    for module in modules:
        for speed, values in content.get(module).items():
            if speed.lower() == 'report':
                parse_report(values, module)
            else:
                for value in values:
                    if not value.get('cmd'):
                        continue
                    item = {
                        'cmd': value.get('cmd'),
                        'output_path': value.get('output_path'),
                        'std_path': value.get('std_path'),
                        'banner': str(value.get('banner')),
                        'module': module,
                        'cmd_type': value.get('cmd_type') if value.get('cmd_type') else 'single',
                        'speed': speed.lower(),
                        'alias': module + "__" + speed.lower() + "__" + str(value.get('banner')).lower(),
                        'chunk': value.get('chunk') if value.get('chunk') else 0,
                    }
                    Commands.objects.create(**item)
    # print(modules)
    return True


