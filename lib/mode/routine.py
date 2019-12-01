import os
import sys

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
from lib.core import utils
from lib.mode import general
from lib.mode import direct
from lib.mode import direct_list
from lib.mode import report
from lib.monitor import backup


def routine_handle(options):
    if options.get('mode') == "report":
        utils.print_load("Running with report mode")
        report.handle(options)
        return

    utils.print_target(options.get('TARGET'))
    # move last result to monitor path and add options to compare
    # only do monitor on general mode
    if options.get('MONITOR') and options['MODE'] == "general":
        options['COMPARE_PATH'] = backup.init_backup(options)
        utils.print_good("Created compare workspaces: {0}".format(
            options['COMPARE_PATH']))

    # really start the routine
    if options['MODE'] == "general":
        general.handle(options)

    elif options['MODE'] == "direct":
        direct.handle(options)

    elif options['MODE'] == "direct_list":
        direct_list.handle(options)

