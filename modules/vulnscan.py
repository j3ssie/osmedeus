import os, time
from core import execute
from core import slack
from core import utils

class VulnScan(object):
    ''' Scanning vulnerable service based version '''
    def __init__(self, options):
        utils.print_banner("Vulnerable Scanning")
        utils.make_directory(options['WORKSPACE'] + '/vulnscan')
        self.module_name = self.__class__.__name__
        self.options = options

        if utils.resume(self.options, self.module_name):
            utils.print_info(
                "It's already done. use '-f' options to force rerun the module")
            return
        self.is_direct = utils.is_direct_mode(options, require_input=True)

        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done Vulnerable Scanning for {0}'.format(self.options['TARGET'])
        })
        self.initial()
        utils.just_waiting(self.options, self.module_name)
        # self.conclude()
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done Vulnerable Scanning for {0}'.format(self.options['TARGET'])
        })

    def initial(self):
        self.nmap_vuln()

    def nmap_vuln(self):
        utils.print_good('Starting Nmap VulnScan')

        if self.is_direct:
            ip_list = utils.just_read(self.is_direct).splitlines()
            ip_list = list(set([ip for ip in ip_list if ip != 'N/A']))
        else:
            main_json = utils.reading_json(utils.replace_argument(
                self.options, '$WORKSPACE/$COMPANY.json'))
            main_json['Modules'][self.module_name] = []

            if self.options['SPEED'] == 'slow':
                ip_list = [x.get("IP")
                        for x in main_json['Subdomains'] if x.get("IP") is not None] + main_json['IP Space']

            elif self.options['SPEED'] == 'quick':
                ip_list = [x.get("IP")
                        for x in main_json['Subdomains'] if x.get("IP") is not None]
            ip_list = list(set([ip for ip in ip_list if ip != 'N/A']))

            if self.options['DEBUG'] == 'True':
                ip_list = list(ip_list)[:5]

        # Scan every 5 IP at time Increse if you want
        for part in utils.chunks(ip_list, 2):
            for ip in part:
                cmd = 'sudo nmap -T4 -Pn -n -sSV -p- {0} --script $PLUGINS_PATH/vulners.nse --oA $WORKSPACE/vulnscan/{0}-nmap'.format(
                    ip.strip())

                cmd = utils.replace_argument(self.options, cmd)
                output_path = utils.replace_argument(
                    self.options, '$WORKSPACE/vulnscan/{0}-nmap.nmap'.format(ip.strip()))
                std_path = utils.replace_argument(
                    self.options, '$WORKSPACE/vulnscan/std-{0}-nmap.std'.format(ip.strip()))
                execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

            # check if previous task done or not every 30 second
            while not utils.checking_done(self.options, module=self.module_name):
                time.sleep(60)

        #just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)

    # def conclude(self):
    #     #### Create beautiful HTML report for masscan
    #     cmd = "xsltproc -o $WORKSPACE/portscan/final-$OUTPUT.html $PLUGINS_PATH/nmap-bootstrap.xsl $WORKSPACE/vulnscan/{0}-nmap"
    #     cmd = utils.replace_argument(self.options, cmd)
    #     output_path = utils.replace_argument(
    #         self.options, '$WORKSPACE/portscan/final-$OUTPUT.html')
    #     std_path = utils.replace_argument(
    #         self.options, '')
    #     execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)
