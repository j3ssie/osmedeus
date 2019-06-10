import os
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
        utils.make_directory(options['WORKSPACE'] + '/subdomain')
        self.module_name = self.__class__.__name__
        self.options = options

        if utils.resume(self.options, self.module_name):
            utils.print_info(
                "It's already done. use '-f' options to force rerun the module")
            return
        
        self.is_direct = utils.is_direct_mode(options, require_input=True)

        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Port Scanning for {0}'.format(self.options['TARGET'])
        })

        self.initial()

        utils.just_waiting(self.options, self.module_name, seconds=60)
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

        # check if direct input is file or just single string
        if self.is_direct:
            if utils.not_empty_file(self.is_direct):
                cmd = '$PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -t A -o S -w $WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt $INPUT_LIST'
            # just return if direct input is just a string
            else:
                return

        else:
            final_ip = utils.replace_argument(
                self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')

            if utils.not_empty_file(final_ip):
                return
            cmd = '$PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -t A -o S -w $WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt $WORKSPACE/subdomain/final-$OUTPUT.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt')

        execute.send_cmd(self.options, cmd, '', '', self.module_name)
        utils.just_waiting(self.options, self.module_name, seconds=5)

        # matching IP with subdomain
        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))

        # get ips from amass stuff
        ips = []
        if self.is_direct:
            if self.options.get("INPUT_LIST"):
                ips.extend(utils.extract_ip(self.options.get('INPUT_LIST')))

        if utils.not_empty_file(output_path):
            data = utils.just_read(output_path).splitlines()
            for line in data:
                if " A " in line:
                    subdomain = line.split('. A ')[0]
                    ip = line.split('. A ')[1]
                    ips.append(str(ip))
                    for i in range(len(main_json['Subdomains'])):
                        if subdomain == main_json['Subdomains'][i]['Domain']:
                            main_json['Subdomains'][i]['IP'] = ip

        final_ip = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')

        utils.just_write(final_ip, "\n".join(ips))
        utils.just_write(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)

    def prepare_input(self):
        if self.is_direct:
            if utils.not_empty_file(self.is_direct):
                ip_file = utils.replace_argument(
                    self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')
                # print(ip_file)
                ip_list = utils.just_read(ip_file).splitlines()
                ip_list = list(set([ip for ip in ip_list if ip != 'N/A']))
            else:
                ip_list = utils.resolve_input(self.is_direct)
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

            ip_list = set([ip for ip in ip_list if ip != 'N/A'])

        return ip_list

    def masscan(self):
        utils.print_good('Starting masscan')
        time.sleep(1)

        ip_list = self.prepare_input()

        if self.is_direct:
            if type(ip_list) == list:
                if self.options['DEBUG'] == "True":
                    utils.print_info("just testing 5 first host")
                    ip_list = list(ip_list)[:5]

                utils.just_write(utils.replace_argument(self.options, '$WORKSPACE/subdomain/IP-$TARGET.txt'), "\n".join(ip_list))

                # print(ip_list)
                time.sleep(1)

                cmd = "sudo masscan --rate 10000 -p0-65535 -iL $WORKSPACE/subdomain/IP-$TARGET.txt -oX $WORKSPACE/portscan/$OUTPUT-masscan.xml --wait 0"
            else:
                cmd = "sudo masscan --rate 10000 -p0-65535 {0} -oX $WORKSPACE/portscan/$OUTPUT-masscan.xml --wait 0".format(ip_list)
        else:
            cmd = "sudo masscan --rate 10000 -p0-65535 -iL $WORKSPACE/subdomain/final-IP-$TARGET.txt -oX $WORKSPACE/portscan/$OUTPUT-masscan.xml --wait 0"

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.xml')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/std-$OUTPUT-masscan.std')
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)


    # Create beautiful HTML report for masscan
    def create_html_report(self):
        cmd = "xsltproc -o $WORKSPACE/portscan/final-$OUTPUT.html $PLUGINS_PATH/nmap-stuff/nmap-bootstrap.xsl $WORKSPACE/portscan/$OUTPUT-masscan.xml"

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/final-$OUTPUT.html')
        std_path = utils.replace_argument(
            self.options, '')
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

    def parsing_to_csv(self):
        masscan_xml = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.xml')
        if not utils.not_empty_file(masscan_xml):
            return

        cmd = "python3 $PLUGINS_PATH/nmap-stuff/masscan_xml_parser.py -f $WORKSPACE/portscan/$OUTPUT-masscan.xml -csv $WORKSPACE/portscan/$OUTPUT-masscan.csv"

        cmd = utils.replace_argument(self.options, cmd)
        csv_output = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.csv')
        execute.send_cmd(self.options, cmd, csv_output,
                        '', self.module_name)
        
        time.sleep(2)

        # csv beatiful
        if not utils.not_empty_file(csv_output):
            return 

        cmd = "cat $WORKSPACE/portscan/$OUTPUT-masscan.csv | csvlook --no-inference | tee $WORKSPACE/portscan/$OUTPUT-masscan-summary.txt"
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/$OUTPUT-masscan-summary.txt')
        execute.send_cmd(self.options, utils.replace_argument(self.options, cmd), output_path,'', self.module_name)
        
        time.sleep(2)
        # re-screeenshot the result with open port
        csv_data = utils.just_read(csv_output)
        self.screenshots(csv_data)
    
    def screenshots(self, csv_data):
        # add http:// and https:// prefix to domain
        if csv_data:
            result = []
            for line in csv_data.splitlines()[1:]:
                # print(line)
                host = line.split(',')[0]
                port = line.split(',')[3]
                result.append("http://" + host + ":" + port)
                result.append("https://" + host + ":" + port)

            utils.just_write(utils.replace_argument(
                self.options, '$WORKSPACE/portscan/$OUTPUT-hosts.txt'), "\n".join(result))
            
            # screenshots with gowitness
            utils.make_directory(self.options['WORKSPACE'] + '/portscan/screenshoots-massscan/')
            
            cmd = "$GO_PATH/gowitness file -s $WORKSPACE/portscan/$OUTPUT-hosts.txt -t 30 --log-level fatal --destination  $WORKSPACE/portscan/screenshoots-massscan/ --db $WORKSPACE/portscan/screenshoots-massscan/gowitness.db"

            execute.send_cmd(self.options, utils.replace_argument(
                self.options, cmd), '', '', self.module_name)
            utils.just_waiting(self.options, self.module_name, seconds=10)

            cmd = "$GO_PATH/gowitness generate -n $WORKSPACE/portscan/$OUTPUT-masscan-screenshots.html  --destination  $WORKSPACE/portscan/screenshoots-massscan/ --db $WORKSPACE/portscan/screenshoots-massscan/gowitness.db"

            html_path = utils.replace_argument(
                self.options, "$WORKSPACE/portscan/$OUTPUT-masscan-screenshots.html")
            execute.send_cmd(self.options, utils.replace_argument(
                self.options, cmd), html_path, '', self.module_name)


    def conclude(self):
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.xml')
        
        if not utils.not_empty_file(output_path):
            return

        self.create_html_report()
        self.parsing_to_csv()

        # parsing masscan xml
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
        
        # update the main json
        for i in range(len(main_json['Subdomains'])):
            ip = main_json['Subdomains'][i].get('IP')
            if ip != "N/A" and ip in masscan_json.keys():
                main_json['Subdomains'][i]['Ports'] = masscan_json.get(ip)

        # just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)

        utils.just_write(utils.replace_argument(
                    self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)






