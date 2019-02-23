import os, time
from core import execute
from core import utils

class SSLScan(object):
    """docstring for SslScan"""
    def __init__(self, options):
        utils.print_banner("SSL Scanning")
        utils.make_directory(options['env']['WORKSPACE'] + '/ssl/')
        self.module_name = self.__class__.__name__
        self.options = options
        self.initial()
        utils.just_waiting(self.module_name)
        self.conclude()

    def initial(self):
        self.testssl()

    def testssl(self):
        utils.print_good('Starting testssl')
        if self.options['speed'] == 'slow':
            cmd = 'bash $PLUGINS_PATH/testssl.sh/testssl.sh --parallel --append --logfile $WORKSPACE/ssl/$TARGET-testssl.txt --file $WORKSPACE/subdomain/final-$OUTPUT.txt'
        elif self.options['speed'] == 'quick':
            cmd = 'bash $PLUGINS_PATH/testssl.sh/testssl.sh --parallel --append --logfile $WORKSPACE/ssl/$TARGET-testssl.txt $TARGET'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/ssl/$TARGET-testssl.txt')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/ssl/std-$TARGET-testssl.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)


    #update the main json file
    def conclude(self):
        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))
        main_json['Modules'][self.module_name] = utils.checking_done(module=self.module_name, get_json=True)

        #write that json again
        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)
            
        #logging
        logfile=utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(logfile)
        utils.print_banner("{0} Done".format(self.module_name))

