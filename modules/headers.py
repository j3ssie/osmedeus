import os, glob
import time
from core import execute
from core import slack
from core import utils


class HeadersScan(object):
    """docstring for Headers Scan"""
    def __init__(self, options):
        utils.print_banner("Headers Scanning")
        utils.make_directory(options['WORKSPACE'] + '/headers')
        utils.make_directory(options['WORKSPACE'] + '/headers/details')
        self.module_name = self.__class__.__name__
        self.options = options
        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Headers Scanning for {0}'.format(self.options['TARGET'])
        })
        self.initial()
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Headers Scanning for {0}'.format(self.options['TARGET'])
        })

    def initial(self):
        if self.observatory():
            utils.just_waiting(self.module_name)
            self.conclude()

    def observatory(self):
        utils.print_good('Starting observatory')

        if self.options['SPEED'] == 'quick':
            utils.print_good('Skipping {0} in quick mode'.format(self.module_name))
            return None
    
        domain_file = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')
        with open(domain_file, 'r') as d:
            domains = d.read().splitlines()

        # for domain in domains:
        for part in list(utils.chunks(domains, 10)):
            for domain in part:
                cmd = 'observatory {0} --format=json -z --attempts 5 | tee $WORKSPACE/headers/details/$TARGET-observatory.json'.format(
                    domain)

                cmd = utils.replace_argument(self.options, cmd)
                output_path = utils.replace_argument(
                    self.options, '$WORKSPACE/headers/details/$TARGET-observatory.json')
                execute.send_cmd(cmd, output_path, '', self.module_name)

            while not utils.checking_done(module=self.module_name):
                time.sleep(15)
        return True
    #update the main json file

    def conclude(self):
        result_path = utils.replace_argument(
            self.options, '$WORKSPACE/headers/details')

        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))

        #head of csv file
        random_json = utils.reading_json(os.listdir(result_path)[0])
        summary_head = ','.join(random_json.keys()) + ",source\n"

        report_path = utils.replace_argument(
            self.options, '$WORKSPACE/headers/summary-$TARGET.csv')
        with open(report_path, 'w+') as r:
            r.write(summary_head)
        # really_details = {}

        for filename in glob.iglob(result_path + '/**/*.json'):
            details = utils.reading_json(filename)
            summarybody = ",".join([v['pass'] for k, v in details.items()]) + "," + filename + "\n"
            with open(report_path, 'a+') as r:
                r.write(summarybody)
        
        utils.check_output(report_path)
        main_json['Modules'][self.module_name] = {"path": report_path}
        #sending slack std
        cmds_json = utils.checking_done(module=self.module_name, get_json=True)
        slack.slack_std(self.options, cmds_json)

