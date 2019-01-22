import os
from core import execute
from core import utils

class SSLScan(object):
    """docstring for SslScan"""
    def __init__(self, options):
        utils.print_banner("SSL Scanning")
        utils.make_directory(options['env']['WORKSPACE'] + '/ssl/')
        self.options = options
        self.initial()


    def initial(self):
        self.testssl()

    def testssl(self):
        utils.print_good('Starting testssl')
        cmd = 'bash $PLUGINS_PATH/testssl.sh/testssl.sh --parallel --logfile $WORKSPACE/ssl/$TARGET-testssl.txt $TARGET'
        cmd = utils.replace_argument(self.options, cmd)
        utils.print_info("Execute: {0} ".format(cmd))
        execute.run(cmd)
        utils.check_output(self.options, '$WORKSPACE/ssl/$TARGET-testssl.txt')
