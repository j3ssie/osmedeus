import time 
from core import execute
from core import slack
from core import utils


class VulnScan(object):
    ''' Scanning vulnerable service based version '''
    def __init__(self, options):
        utils.print_banner("Vulnerable Scanning")
        utils.make_directory(options['WORKSPACE'] + '/vulnscan')
        utils.make_directory(options['WORKSPACE'] + '/vulnscan/details')
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
        utils.just_waiting(self.options, self.module_name, seconds=120)

        self.conclude()
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done Vulnerable Scanning for {0}'.format(self.options['TARGET'])
        })

    def initial(self):
        ip_list = self.prepare_input()
        if type(ip_list) == list:
            self.nmap_vuln_list(ip_list)
        else:
            self.nmap_single(ip_list)


    def prepare_input(self):
        if self.is_direct:
            # if direct input was file just read it
            if utils.not_empty_file(self.is_direct):
                ip_list = utils.just_read(self.is_direct).splitlines()
            # get input string
            else:
                ip_list = utils.get_domain(self.is_direct).strip()

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

        # utils.print_debug(ip_list)
        return ip_list


    def nmap_single(self, input_target):
        encode_input = utils.strip_slash(input_target)
        cmd = 'sudo nmap --open -T4 -Pn -n -sSV -p- {0} --script $PLUGINS_PATH/nmap-stuff/vulners.nse --oA $WORKSPACE/vulnscan/details/{1}-nmap'.format(
            input_target, encode_input)

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/vulnscan/details/{0}-nmap.nmap'.format(encode_input))
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/vulnscan/details/std-{0}-nmap.std'.format(encode_input))
        execute.send_cmd(self.options, cmd, output_path,
                         std_path, self.module_name)

    def nmap_vuln_list(self, ip_list):
        utils.print_good('Starting Nmap VulnScan')

        # Scan every 2 IP at time Increse if you want
        for part in utils.chunks(ip_list, 2):
            for ip in part:
                encode_input = utils.strip_slash(ip.strip())
                cmd = 'sudo nmap --open -T4 -Pn -n -sSV -p- {0} --script $PLUGINS_PATH/nmap-stuff/vulners.nse --oA $WORKSPACE/vulnscan/details/{1}-nmap'.format(
                    ip.strip(), encode_input)

                cmd = utils.replace_argument(self.options, cmd)
                output_path = utils.replace_argument(
                    self.options, '$WORKSPACE/vulnscan/details/{0}-nmap.nmap'.format(ip.strip()))
                std_path = utils.replace_argument(
                    self.options, '$WORKSPACE/vulnscan/details/std-{0}-nmap.std'.format(ip.strip()))
                execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

            # check if previous task done or not every 30 second
            utils.just_waiting(self.options, self.module_name, seconds=120)

        # just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)

    # parsing the output and screenshot with some new url
    def parsing_to_csv(self):
        nmap_detail_path = utils.replace_argument(self.options, '$WORKSPACE/vulnscan/details')

        # create all csv based on xml file
        for file in utils.list_files(nmap_detail_path, ext='xml'):
            # print(file)
            cmd = "python3 $PLUGINS_PATH/nmap-stuff/nmap_xml_parser.py -f {0} -csv $WORKSPACE/vulnscan/details/$OUTPUT-nmap.csv".format(file)

            cmd = utils.replace_argument(self.options, cmd)
            csv_output = utils.replace_argument(
                self.options, '$WORKSPACE/vulnscan/details/$OUTPUT-nmap.csv')
            execute.send_cmd(self.options, cmd, csv_output,
                            '', self.module_name)

        time.sleep(5)
        # looping through all csv file
        all_csv = "IP,Host,OS,Proto,Port,Service,Product,Service FP,NSE Script ID,NSE Script Output,Notes\n"

        for file in utils.list_files(nmap_detail_path, ext='csv'):
            all_csv += "\n".join(utils.just_read(file).splitlines()[1:])

        csv_summary_path = utils.replace_argument(
            self.options, '$WORKSPACE/vulnscan/summary-$OUTPUT.csv')
        utils.just_write(csv_summary_path, all_csv)

        # beautiful csv look
        cmd = "csvcut -c 1-7 $WORKSPACE/vulnscan/summary-$OUTPUT.csv | csvlook  --no-inference | tee $WORKSPACE/vulnscan/std-$OUTPUT-summary.std"
        
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/portscan/std-$OUTPUT-summary.std')
        execute.send_cmd(self.options, cmd, output_path, '', self.module_name)

        self.screenshots(all_csv)


    def screenshots(self, csv_data):
        utils.print_info("Screenshot again with new port found")
        # add http:// and https:// prefix to domain
        if csv_data:
            result = []
            for line in csv_data.splitlines()[1:]:
                # some output of the script is contain new line
                try:
                    # print(line)
                    host = line.split(',')[0]
                    port = line.split(',')[4]
                    result.append("http://" + host + ":" + port)
                    result.append("https://" + host + ":" + port)
                except:
                    pass

            utils.just_write(utils.replace_argument(
                self.options, '$WORKSPACE/vulnscan/$OUTPUT-hosts.txt'), "\n".join(result))

            # screenshots with gowitness
            utils.make_directory(
                self.options['WORKSPACE'] + '/vulnscan/screenshoots-nmap/')

            cmd = "$GO_PATH/gowitness file -s $WORKSPACE/vulnscan/$OUTPUT-hosts.txt -t 30 --log-level fatal --destination  $WORKSPACE/vulnscan/screenshoots-nmap/ --db $WORKSPACE/vulnscan/screenshoots-nmap/gowitness.db"

            execute.send_cmd(self.options, utils.replace_argument(
                self.options, cmd), '', '', self.module_name)
            utils.just_waiting(self.options, self.module_name, seconds=10)

            cmd = "$GO_PATH/gowitness generate -n $WORKSPACE/vulnscan/$OUTPUT-nmap-screenshots.html  --destination  $WORKSPACE/vulnscan/screenshoots-nmap/ --db $WORKSPACE/vulnscan/screenshoots-nmap/gowitness.db"

            html_path = utils.replace_argument(
                self.options, "$WORKSPACE/vulnscan/$OUTPUT-nmap-screenshots.html")
            execute.send_cmd(self.options, utils.replace_argument(
                self.options, cmd), html_path, '', self.module_name)


    def conclude(self):
        self.parsing_to_csv()

