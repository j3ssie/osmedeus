import os, time
from core import execute
from core import slack
from core import utils

class ScreenShot(object):
    """Screenshot all domain on common service"""
    def __init__(self, options):
        utils.print_banner("ScreenShot the target")
        utils.make_directory(options['WORKSPACE'] + '/screenshot')
        self.module_name = self.__class__.__name__
        self.options = options
        if utils.resume(self.options, self.module_name):
            utils.print_info("Detect is already done. use '-f' options to force rerun the module")
            return
        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start ScreenShot for {0}'.format(self.options['TARGET'])
        })
        self.initial()
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done ScreenShot for {0}'.format(self.options['TARGET'])
        })


    def initial(self):
        self.run()
        utils.just_waiting(self.options, self.module_name, seconds=10)
        #this gonna run after module is done to update the main json
        # self.conclude()

    def run(self):
        commands = execute.get_commands(self.options, self.module_name).get('routines')

        for item in commands:
            utils.print_good('Starting {0}'.format(item.get('banner')))
            #really execute it
            execute.send_cmd(self.options, item.get('cmd'), item.get(
                'output_path'), item.get('std_path'), self.module_name)
            time.sleep(1)

        utils.just_waiting(self.options, self.module_name, seconds=30)
        #just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)
    

    # #update the main json file
    # def conclude(self):
    #     output_path = utils.replace_argument(
    #         self.options, '$WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt')

    #     # matching IP with subdomain
    #     main_json = utils.reading_json(utils.replace_argument(
    #         self.options, '$WORKSPACE/$COMPANY.json'))
    #     with open(output_path, 'r') as i:
    #         data = i.read().splitlines()
    #     ips = []
    #     for line in data:
    #         if " A " in line:
    #             subdomain = line.split('. A ')[0]
    #             ip = line.split('. A ')[1]
    #             ips.append(ip)
    #             for i in range(len(main_json['Subdomains'])):
    #                 if subdomain == main_json['Subdomains'][i]['Domain']:
    #                     main_json['Subdomains'][i]['IP'] = ip

    #     final_ip = utils.replace_argument(
    #         self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')

    #     with open(final_ip, 'w+') as fip:
    #         fip.write("\n".join(str(ip) for ip in ips))

    #     utils.just_write(utils.replace_argument(
    #         self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)
