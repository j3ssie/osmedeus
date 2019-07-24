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
        self.options['CURRENT_MODULE'] = self.module_name
        self.options['SPEED'] = utils.custom_speed(self.options)
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
        utils.print_line()

    def initial(self):
        domains, http_domains_path = self.prepare_input()
        self.dirsearch(http_domains_path)
        self.dirble(http_domains_path)
        self.parsing_ouput()
        self.screenshots()

    def prepare_input(self):
        if self.is_direct:
            # if direct input was file just read it
            if utils.not_empty_file(self.is_direct):
                domains = utils.just_read(self.is_direct).splitlines()
                http_domains_path = self.is_direct
            # get input string
            else:
                domains = [self.is_direct.strip()]
                http_domains_path = utils.reading_json(utils.replace_argument(
                    self.options, '$WORKSPACE/directory/domain-lists.txt'))
                utils.just_write(http_domains_path, "\n".join(domains))
        else:
            http_domains_path = utils.replace_argument(
                self.options, '$WORKSPACE/assets/http-$OUTPUT.txt')
            # if assets module done return it
            if utils.not_empty_file(http_domains_path):
                domains = utils.just_read(http_domains_path).splitlines()
                return domains, http_domains_path

            # matching IP with subdomain
            main_json = utils.reading_json(utils.replace_argument(
                self.options, '$WORKSPACE/$COMPANY.json'))
            domains = [x.get('Domain') for x in main_json['Subdomains']]

            http_domains_path = utils.reading_json(utils.replace_argument(
                self.options, '$WORKSPACE/directory/domain-lists.txt'))
            utils.just_write(http_domains_path, "\n".join(domains))

        return domains, http_domains_path

    # Dirsearch list of target
    def dirsearch(self, domain_path):
        utils.print_good('Starting dirsearch')
        cmd = "python3 $PLUGINS_PATH/dirsearch/dirsearch.py -b -e php,aspx,jsp,swp,swf,zip --wordlist=$PLUGINS_PATH/wordlists/really-quick.txt -x '302,404' --simple-report=$WORKSPACE/directory/quick/$OUTPUT-dirsearch.txt -t 50 -L {0}".format(
            domain_path)

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/quick/$OUTPUT-dirsearch.txt')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/quick/std-$OUTPUT-dirsearch.std')
        execute.send_cmd(self.options, cmd, output_path,
                         std_path, self.module_name)

        utils.just_waiting(self.options, self.module_name, seconds=120)

    # dirble list of target
    def dirble(self, domain_path):
        if self.options['SPEED'] == 'quick':
            utils.print_info('Skip dirble in quick speed')
            return None
        utils.print_good('Starting dirble')
        cmd = './dirble -U {0}  -x ".php,.aspx,.jsp,.swp,.swf,.zip" -t 40 -k -w $PLUGINS_PATH/wordlists/really-quick.txt --output-file $WORKSPACE/directory/full/$OUTPUT-dirble.txt'.format(
            domain_path)

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/full/$OUTPUT-dirble.txt')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/directory/full/std-$OUTPUT-dirble.std')
        execute.send_cmd(self.options, cmd, output_path,
                         std_path, self.module_name)

    # checking if it's done and get all found path
    def parsing_ouput(self):
        utils.print_good('Parsing result found to a file')
        final_result = utils.replace_argument(
            self.options, '$WORKSPACE/directory/$OUTPUT-summary.txt')

        dirsearch_result = utils.replace_argument(
            self.options, '$WORKSPACE/directory/quick/$OUTPUT-dirsearch.txt')
        data = utils.just_read(dirsearch_result)
        if data:
            utils.just_append(final_result, data)

        dirble_result = utils.replace_argument(
            self.options, '$WORKSPACE/directory/full/$OUTPUT-dirble.txt')
        data = utils.just_read(dirble_result)
        if data:
            utils.just_append(final_result, data)

        # final_result
        utils.clean_up(final_result)
        utils.check_output(final_result)

    # screenshots all result found
    def screenshots(self):
        if self.options['SPEED'] == 'quick':
            utils.print_info('Skip screenshot on Dirbrute in quick speed')
            return None
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


