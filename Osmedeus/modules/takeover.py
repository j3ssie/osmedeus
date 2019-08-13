import os, time
from pprint import  pprint

from Osmedeus.core import execute
from Osmedeus.core import slack
from Osmedeus.core import utils 

class TakeOverScanning(object):
    def __init__(self, options):
        utils.print_banner("Scanning for Subdomain TakeOver")
        self.module_name = self.__class__.__name__
        self.options = options
        self.options['CURRENT_MODULE'] = self.module_name
        self.options['SPEED'] = utils.custom_speed(self.options)
        if utils.resume(self.options, self.module_name):
            utils.print_info("It's already done. use '-f' options to force rerun the module")
            return
        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning TakeOver for {0}'.format(self.options['TARGET'])
        })
        self.initial()

        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done Scanning TakeOver for {0}'.format(self.options['TARGET'])
        })

        utils.print_banner("{0} Done".format(self.module_name))


    def initial(self):
        self.run()
        # if self.options['SPEED'] == 'slow':
        #     self.dig_info()

    def run(self):
        commands = execute.get_commands(self.options, self.module_name).get('routines')
        for item in commands:
            utils.print_good('Starting {0}'.format(item.get('banner')))
            #really execute it
            execute.send_cmd(self.options, item.get('cmd'), item.get(
                'output_path'), item.get('std_path'), self.module_name)

        utils.just_waiting(self.options, self.module_name, seconds=20, times=5)
        #just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)

    def dig_info(self):
        utils.print_good('Starting basic Dig')
        utils.make_directory(self.options['WORKSPACE'] + '/screenshot/digs')
        final_subdomains = utils.replace_argument(self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')

        #run command directly instead of run it via module cause there're a lot of command to run
        all_domains = utils.just_read(final_subdomains).splitlines()
        
        if self.options['DEBUG'] == 'True':
            all_domains = all_domains[:10]

        custom_logs = {"module": self.module_name, "content": []}
        for part in list(utils.chunks(all_domains, 5)):
            for domain in part:
                cmd = utils.replace_argument(
                    self.options, 'dig all {0} | tee $WORKSPACE/screenshot/digs/{0}.txt'.format(domain))
                
                output_path =  utils.replace_argument(self.options, 'tee $WORKSPACE/screenshot/digs/{0}.txt'.format(domain))
                execute.send_cmd(self.options, cmd, '', '', self.module_name, True)
                # time.sleep(0.5)

                custom_logs['content'].append(
                    {"cmd": cmd, "std_path": '', "output_path": output_path, "status": "Done"})
            #just wait couple seconds and continue but not completely stop the routine
            time.sleep(5)
            
        print(custom_logs)
        #submit a log
        utils.print_info('Update activities log')
        utils.update_activities(self.options, str(custom_logs))
        utils.print_line()
