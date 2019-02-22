import os
from core import execute
from core import utils

class CorsScan(object):
    """docstring for PortScan"""
    def __init__(self, options):
        utils.print_banner("CORS Scanning")
        utils.make_directory(options['env']['WORKSPACE'] + '/cors/')
        self.module_name = self.__class__.__name__
        self.options = options
        self.initial()
        utils.just_waiting(self.module_name)
        self.conclude()


    def initial(self):
        self.corstest()

    def corstest(self):
        utils.print_good('Starting CORS')
        cmd = 'python2.7 $PLUGINS_PATH/CORStest/corstest.py $WORKSPACE/subdomain/final-$OUTPUT.txt | tee $WORKSPACE/cors/$TARGET-corstest.txt'
        
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/cors/$TARGET-corstest.txt')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/cors/std-$TARGET-corstest.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)

    #update the main json file
    def conclude(self):
        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))
        main_json['Modules'][self.module_name] = utils.checking_done(module=self.module_name, get_json=True)

        #write that json again
        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)        
        #logging
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(logfile)

