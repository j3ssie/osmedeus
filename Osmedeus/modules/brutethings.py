import os, socket

from Osmedeus.core import execute
from Osmedeus.core import utils

class BruteThings(object):
    """docstring for BruteThings"""
    def __init__(self, options):
        utils.make_directory(options['WORKSPACE'] + '/bruteforce/')
        self.options = options
        self.options['CURRENT_MODULE'] = self.module_name
        self.options['SPEED'] = utils.custom_speed(self.options)
        
        if utils.resume(self.options, self.module_name):
            utils.print_info("It's already done. use '-f' options to force rerun the module")
            return

        if self.options['SPEED'] == 'slow':
            self.routine()
        elif self.options['SPEED'] == 'quick':
            utils.print_good("Skipping for quick speed")



    # normal routine
    def initial(self):
        self.brutespray()

    def brutespray(self):
        utils.print_good('Starting brutespray')
        cmd = 'python $PLUGINS_PATH/brutespray/brutespray.py --file $WORKSPACE/vulnscan/$TARGET-nmap.xml --threads 5 --hosts 5 -o $WORKSPACE/bruteforce/$OUTPUT/'
        cmd = utils.replace_argument(self.options, cmd)
        utils.print_info("Execute: {0} ".format(cmd))
        execute.run(cmd)
        utils.check_output(self.options, '$WORKSPACE/bruteforce/$OUTPUT/')
