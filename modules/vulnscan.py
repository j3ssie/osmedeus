import os, time
from core import execute
from core import utils

class VulnScan(object):
    ''' Scanning vulnerable service based version '''
    def __init__(self, options):
        utils.print_banner("Vulnerable Scanning")
        utils.make_directory(options['WORKSPACE'] + '/vulnscan')
        self.module_name = self.__class__.__name__
        self.options = options
        self.initial()
        utils.just_waiting(self.module_name)
        self.conclude()

    def initial(self):
        self.nmap_vuln()

    def nmap_vuln(self):
        utils.print_good('Starting Nmap VulnScan')
        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))
        main_json['Modules'][self.module_name] = []

        if self.options['speed'] == 'slow':
            ip_list = [x.get("IP")
                       for x in main_json['Subdomains']] + main_json['IP Space']

        elif self.options['speed'] == 'quick':
            ip_list = [x.get("IP") for x in main_json['Subdomains']]

        # Scan every 5 IP at time Increse if you want
        for part in list(utils.chunks(ip_list, 5)):
            for ip in part:
                cmd = 'sudo nmap -T4 -Pn -n -sSV -p- {0} --script vulners --oA $WORKSPACE/vulnscan/$OUTPUT-nmap'.format(ip)

                cmd = utils.replace_argument(self.options, cmd)
                output_path = utils.replace_argument(
                    self.options, '$WORKSPACE/vulnscan/{0}-nmap.nmap'.format(ip))
                std_path = utils.replace_argument(
                    self.options, '$WORKSPACE/vulnscan/std-{0}-nmap.std'.format(ip))
                execute.send_cmd(cmd, output_path, std_path, self.module_name)

            # check if previous task done or not every 30 second
            while not utils.checking_done(module=self.module_name):
                time.sleep(20)

            # update main json
            main_json['Modules'][self.module_name] += utils.checking_done(
                module=self.module_name, get_json=True)

    def conclude(self):
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(logfile)
    
    # def create_html(self):
    #     utils.print_good('Create beautify HTML report')
    #     cmd = 'xsltproc -o $WORKSPACE/vulnscan/$OUTPUT.html $PLUGINS_PATH/nmap-bootstrap.xsl $WORKSPACE/vulnscan/$OUTPUT-nmap.xml'
    #     cmd = utils.replace_argument(self.options, cmd)
    #     utils.print_info("Execute: {0} ".format(cmd))
    #     execute.run(cmd)
    #     utils.check_output(self.options, '$WORKSPACE/vulnscan/$TARGET.html')
