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
        if utils.resume(self.options, self.module_name):
            utils.print_info("Detect is already done. use '-f' options to force rerun the module")
            return
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
            utils.just_waiting(self.options, self.module_name)
            try:
                self.conclude()
            except:
                utils.print_bad("Something wrong with conclude for {0}".format(self.module_name))

    def observatory(self):
        utils.print_good('Starting observatory')

        if self.options['SPEED'] == 'quick':
            utils.print_good('Skipping {0} in quick mode'.format(self.module_name))
            return None
    
        domain_file = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')
        with open(domain_file, 'r') as d:
            domains = d.read().splitlines()

        if self.options['DEBUG'] == "True":
            utils.print_info("Only get 30 target debug mode")
            domains = domains[:10]

        for part in list(utils.chunks(domains, 10)):
            for domain in part:
                cmd = 'observatory -q {0} --format=json -z --attempts 10 | tee $WORKSPACE/headers/details/{0}-observatory.json'.format(
                    domain.strip())
                cmd = utils.replace_argument(self.options, cmd)
                execute.send_cmd(self.options, cmd, '', '', self.module_name)

            while not utils.checking_done(module=self.module_name):
                time.sleep(15)
        return True
    #update the main json file
    def conclude(self):
        result_path = utils.replace_argument(
            self.options, '$WORKSPACE/headers/details')

        report_path = utils.replace_argument(
            self.options, '$WORKSPACE/headers/summary-$TARGET.csv')

        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))

        #head of csv file
        random_json = utils.reading_json(
            result_path + "/" + os.listdir(result_path)[0])

        summary_head = "domain," + ','.join(random_json.keys()) + ",score,details\n"

        with open(report_path, 'w+') as r:
            r.write(summary_head)

        for filename in os.listdir(result_path):
            real_path = result_path + "/" + filename
            details = utils.reading_json(real_path)
            if details:
                summarybody = filename.replace('-observatory.json', '')
                score = 100
                for k, v in details.items():
                    if "not-implemented" in v.get('result'):
                        summarybody += ',' + "Not Implement"
                    elif v.get('pass') == True:
                        summarybody += ',' + "Pass"
                    else:
                        summarybody += ',' + "Fail"

                    score += int(v.get('score_modifier'))

                #if score is below zero just make it 0 like
                if score < 0:
                    score = 0
                
                summarybody += ',' + str(score) + ',' + real_path + "\n"
                with open(report_path, 'a+') as r:
                    r.write(summarybody)
