import os, time
from core import execute
from core import utils

class IPSpace(object):
    ''' Scanning vulnerable service based version '''
    def __init__(self, options):
        utils.print_banner("IP Discovery")
        utils.make_directory(options['env']['WORKSPACE'] + '/ipspace')
        self.module_name = self.__class__.__name__
        self.options = options

        self.initial()
        utils.just_waiting(self.module_name)
        self.conclude()

    def initial(self):
        self.ipOinst()

    def ipOinst(self):
        utils.print_good('Starting IPOinst')
        cmd = '$PLUGINS_PATH/IPOsint/ip-osint.py -t $COMPANY -o $WORKSPACE/ipspace/$OUTPUT-ipspace.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/ipspace/$OUTPUT-ipspace.txt')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/ipspace/std-$OUTPUT-ipspace.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)


    #update the main json file
    def conclude(self):
        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))
        main_json['Modules'][self.module_name] = utils.checking_done(module=self.module_name, get_json=True)

        ips_file = utils.replace_argument(self.options, '$WORKSPACE/ipspace/$OUTPUT-ipspace.txt')
        with open(ips_file, 'r') as s:
            ips = s.read().splitlines()
        main_json['IP Space'] = ips

        #write that json again
        utils.just_write(utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json')), main_json, is_json=True)
        
        #logging
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(logfile)

