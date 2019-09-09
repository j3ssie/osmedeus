import os
import sys

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
from lib.client import helpers
from lib.core import utils
from lib.sender import send
from lib.sender import initial

from lib.reporter import summaries
from lib.reporter import listws
from lib.reporter import paths
from lib.reporter import exports


def parse_options(options):
    options = utils.lower_dict_keys(options)
    options['workspace'] = utils.get_ws(options.get('raw_target', None))
    if not options['workspace']:
        return False
    arguments = initial.get_workspace_info(options)
    options = {**options, **arguments}
    return utils.upper_dict_keys(options)


def handle(options):
    # parsing some options
    options = utils.upper_dict_keys(options)
    report_type = options.get('REPORT')
    if report_type == 'hh':
        helpers.report_help()
        return
    if utils.any_in(report_type,  ['ls', 'list']):
        listws.show(options)
        return

    # parsing info about workspace
    # print(options)
    options = parse_options(options)
    if not options:
        helpers.report_help()
        return

    report_type = options.get('REPORT')
    # print(options)

    if utils.any_in(report_type, ['sum', 'summary']):
        summaries.show(options)

    if utils.any_in(report_type, ['ex', 'export', 'exp']):
        exports.show(options)

    if utils.any_in(report_type, ['path', 'pa']):
        paths.show(options)

    if utils.any_in(report_type, ['full', 'f']):
        paths.show(options, get_content=True)

