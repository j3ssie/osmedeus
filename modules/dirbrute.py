import os
import time
from core import execute
from core import slack
from core import utils


class DirBrute(object):
    """docstring for DirbBrute"""
    def __init__(self, options):
        utils.print_banner("Scanning Directory")
        utils.make_directory(options['WORKSPACE'] + '/directory')
        utils.make_directory(options['WORKSPACE'] + '/directory/quick')
        utils.make_directory(options['WORKSPACE'] + '/directory/full')
        self.module_name = self.__class__.__name__
        self.options = options
        if utils.resume(self.options, self.module_name):
            utils.print_info("It's already done. use '-f' options to force rerun the module")
            return

        self.is_direct = utils.is_direct_mode(options, require_input=True)


        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning Directory for {0}'.format(self.options['TARGET'])
        })
        self.initial()
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning Directory for {0}'.format(self.options['TARGET'])
        })

    def initial(self):
        domains = self.prepare_input()
        self.dirb_all(domains)
        self.parsing_ouput()
        self.screenshots()

    def prepare_input(self):
        if self.is_direct:
            # if direct input was file just read it
            if utils.not_empty_file(self.is_direct):
                domains = utils.just_read(self.is_direct).splitlines()
            # get input string
            else:
                domains = [self.is_direct.strip()]
        else:
            # matching IP with subdomain
            main_json = utils.reading_json(utils.replace_argument(
                self.options, '$WORKSPACE/$COMPANY.json'))
            domains = [x.get('Domain') for x in main_json['Subdomains']]
        
        return domains

    def dirb_all(self, domains):
        for domain in domains:
            # passing domain to content directory tools
            self.dirsearch(domain.strip())
            self.wfuzz(domain.strip())
            self.gobuster(domain.strip())

            # just wait couple seconds and continue but not completely stop the routine
            time.sleep(60)
            
        # brute with 3 tools
        utils.force_done(self.options, self.module_name)
        # just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)


    # checking if it's done and get all found path
    def parsing_ouput(self):
        utils.print_good('Parsing result found to a file')
        final_result = utils.replace_argument(
            self.options, '$WORKSPACE/directory/$OUTPUT-summary.txt')

        # dirsearch part
        dirsearch_files = utils.list_files(self.options['WORKSPACE'] +'/directory/quick', pattern='*-dirsearch.txt')
        for file in dirsearch_files:
            data = utils.just_read(file)
            if data:
                utils.just_append(final_result, data)

        # wfuzz part
        wfuzz_files =  utils.list_files(
            self.options['WORKSPACE'] + '/directory/quick', pattern='*-wfuzz.json')
        for file in wfuzz_files:
            data_json = utils.reading_json(file)
            if data_json:
                data = "\n".join([x.get("url") for x in data_json])
                utils.just_append(final_result, data)

        # final_result
        utils.clean_up(final_result)
        utils.check_output(final_result)

    # screenshots all result found
    def screenshots(self):
        utils.print_good('Starting Screenshot from found result')
        final_result = utils.replace_argument(self.options, '$WORKSPACE/directory/$OUTPUT-summary.txt')
        if utils.not_empty_file(final_result):
            # screenshot found path at the end
            cmd = "cat {0} | $GO_PATH/aquatone -threads 20 -out $WORKSPACE/directory/$OUTPUT-screenshots".format(
                final_result)

            cmd = utils.replace_argument(self.options, cmd)
            std_path = utils.replace_argument(self.options, '$WORKSPACE/directory/$OUTPUT-screenshots/std-aquatone_report.std')
            output_path = utils.replace_argument(self.options, '$WORKSPACE/directory/$OUTPUT-screenshots/aquatone_report.html')
            execute.send_cmd(self.options, cmd, std_path, output_path, self.module_name)

            if utils.not_empty_file(output_path):
                utils.check_output(output_path)

    # Just boring replicate similar command for content directory tools
    def dirsearch(self, domain):
        strip_domain = utils.get_domain(domain)

        utils.print_good('Starting dirsearch')
        cmd = "python3 $PLUGINS_PATH/dirsearch/dirsearch.py -b -e php,zip,aspx,js --wordlist=$PLUGINS_PATH/wordlists/really-quick.txt -x '302,404' --simple-report=$WORKSPACE/directory/quick/{1}-dirsearch.txt -t 50 -u {0}".format(domain, strip_domain)

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/quick/{0}-dirsearch.txt'.format(strip_domain))
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/quick/std-{0}-dirsearch.std'.format(strip_domain))
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

    def wfuzz(self, domain):
        strip_domain = utils.get_domain(domain)

        utils.print_good('Starting wfuzz')
        cmd = "wfuzz -f $WORKSPACE/directory/quick/{1}-wfuzz.json,json -c -w $PLUGINS_PATH/wordlists/quick-content-discovery.txt -t 100 --sc 200,307 -u '{0}/FUZZ' | tee $WORKSPACE/directory/quick/std-{1}-wfuzz.std".format(domain, strip_domain)

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/quick/{0}-wfuzz.json'.format(strip_domain))
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/quick/std-{0}-wfuzz.std'.format(strip_domain))
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

    def gobuster(self, domain):
        utils.print_good('Starting gobuster')
        if self.options['SPEED'] != 'slow':
            utils.print_good("Skipping gobuster in quick mode")
            return

        strip_domain = utils.get_domain(domain)
        cmd = '$GO_PATH/gobuster dir -k -q -e -fw -x php,jsp,aspx,html,json -w $PLUGINS_PATH/wordlists/dir-all.txt -t 100 -o $WORKSPACE/directory/{1}-gobuster.txt -s 200,301,307 -u "{0}" '.format(
            domain, strip_domain)

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/full/{0}-gobuster.json'.format(domain))
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/full/std-{0}-gobuster.std'.format(domain))
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name, nolog=True)

