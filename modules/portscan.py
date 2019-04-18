import os
import glob
import socket
import time
import json
import xml.etree.ElementTree as ET
from core import execute
from core import slack
from core import utils

class PortScan(object):
    """docstring for PortScan"""

    def __init__(self, options):
        utils.print_banner("Port Scanning")
        utils.make_directory(options['WORKSPACE'] + '/portscan')
        self.module_name = self.__class__.__name__
        self.options = options
        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Port Scanning for {0}'.format(self.options['TARGET'])
        })
        self.initial()

        utils.just_waiting(self.options, self.module_name, seconds=60)
        # self.create_ip_result()
        self.conclude()
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Port Scanning for {0}'.format(self.options['TARGET'])
        })

    def initial(self):
        self.create_ip_result()
        self.masscan()

    # just for the masscan
    def create_ip_result(self):
        utils.print_good('Create IP for list of domain result')
        final_ip = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')
        
        # if utils.not_empty_file(final_ip):
        #     return

        cmd = '$PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -t A -o S -w $WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt $WORKSPACE/subdomain/final-$OUTPUT.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt')
        execute.send_cmd(self.options, cmd, '', '', self.module_name)

        utils.just_waiting(self.options, self.module_name, seconds=5)

        # matching IP with subdomain
        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))
        with open(output_path, 'r') as i:
            data = i.read().splitlines()
        ips = []
        for line in data:
            if " A " in line:
                subdomain = line.split('. A ')[0]
                ip = line.split('. A ')[1]
                ips.append(ip)
                for i in range(len(main_json['Subdomains'])):
                    if subdomain == main_json['Subdomains'][i]['Domain']:
                        main_json['Subdomains'][i]['IP'] = ip

        final_ip = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')

        with open(final_ip, 'w+') as fip:
            fip.write("\n".join(str(ip) for ip in ips))

        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)

    def masscan(self):
        utils.print_good('Starting masscan')
        time.sleep(1)

        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))
        main_json['Modules'][self.module_name] = []

        if self.options['SPEED'] == 'slow':
            ip_list = [x.get("IP")
                       for x in main_json['Subdomains'] if x.get("IP") is not None] + main_json['IP Space']

        elif self.options['SPEED'] == 'quick':
            ip_list = [x.get("IP")
                       for x in main_json['Subdomains'] if x.get("IP") is not None]

        ip_list = set([ip for ip in ip_list if ip != 'N/A'])

        if self.options['DEBUG'] == "True":
            utils.print_info("just testing 5 first host")
            ip_list = list(ip_list)[:5]

        # print(ip_list)

        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/IP-$TARGET.txt'), "\n".join(ip_list))

        # print(ip_list)
        time.sleep(1)

        # cmd = "sudo masscan --rate 10000 -p0-65535 -iL $WORKSPACE/subdomain/IP-$TARGET.txt -oJ $WORKSPACE/portscan/$OUTPUT-masscan.json --wait 0"
        cmd = "sudo masscan --rate 1000 -p0-65535 -iL $WORKSPACE/subdomain/IP-$TARGET.txt -oX $WORKSPACE/portscan/$OUTPUT-masscan.xml --wait 0"

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.xml')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/std-$OUTPUT-masscan.std')
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)
    
    #Create beautiful HTML report for masscan
    def create_html_report(self):
        cmd = "xsltproc -o $WORKSPACE/portscan/final-$OUTPUT.html $PLUGINS_PATH/nmap-bootstrap.xsl $WORKSPACE/portscan/$OUTPUT-masscan.xml"
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/final-$OUTPUT.html')
        std_path = utils.replace_argument(
            self.options, '')
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

    def conclude(self):
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.xml')
        
        if not utils.not_empty_file(output_path):
            return

        self.create_html_report()

        #parsing masscan xml
        tree = ET.parse(output_path)
        root = tree.getroot()
        masscan_json = {}
        for host in root.iter('host'):
            ip = host[0].get('addr')
            ports = [(str(x.get('portid')) + "/" + str(x.get('protocol')))
                     for x in host[1]]
            masscan_json[ip] = ports

        main_json = utils.reading_json(utils.replace_argument(
                self.options, '$WORKSPACE/$COMPANY.json'))
        
        #update the main json
        for i in range(len(main_json['Subdomains'])):
            ip = main_json['Subdomains'][i].get('IP')
            if ip != "N/A" and ip in masscan_json.keys():
                main_json['Subdomains'][i]['Ports'] = masscan_json.get(ip)


        utils.just_write(utils.replace_argument(
                    self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)






