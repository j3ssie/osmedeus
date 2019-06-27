import os
import time
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
        self.options['CURRENT_MODULE'] = self.module_name
        self.options['SPEED'] = utils.custom_speed(self.options)
        if utils.resume(self.options, self.module_name):
            utils.print_info("It's already done. use '-f' options to force rerun the module")
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
        utils.print_line()

    # check if this was run on subdomain module or direct mode from screenshot
    def check_direct(self):
        all_subdomain = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')

        if utils.not_empty_file(all_subdomain):
            return False

        self.is_direct = utils.is_direct_mode(self.options, require_input=True)
        return True

    def initial(self):
        if self.check_direct():
            self.screenshots(self.is_direct)
        else:
            self.run()
        utils.just_waiting(self.options, self.module_name, seconds=10)
        # this gonna run after module is done to update the main json
        # self.conclude()

    def run(self):
        commands = execute.get_commands(self.options, self.module_name).get('routines')

        for item in commands:
            utils.print_good('Starting {0}'.format(item.get('banner')))
            # really execute it
            execute.send_cmd(self.options, item.get('cmd'), item.get(
                'output_path'), item.get('std_path'), self.module_name)
            time.sleep(1)

        utils.just_waiting(self.options, self.module_name, seconds=30)
        # just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)

    def screenshots(self, input_file):
        if not utils.not_empty_file(input_file):
            return False

        data = utils.just_read(input_file).splitlines()
        self.aquatone(input_file)
        self.gowithness(data)

    def aquatone(self, input_file):
        cmd = "cat {0} | $GO_PATH/aquatone -threads 20 -out $WORKSPACE/screenshot/$OUTPUT-aquatone".format(input_file)

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, "$WORKSPACE/screenshot/$OUTPUT-aquatone/aquatone_report.html")
        std_path = utils.replace_argument(
            self.options, "$WORKSPACE/screenshot/std-$OUTPUT-aquatone.std")

        execute.send_cmd(self.options, cmd, output_path,
                         std_path, self.module_name)

    def gowithness(self, data):
        # add http:// and https:// prefix to domain
        domains = []
        utils.make_directory(
            self.options['WORKSPACE'] + '/screenshot/screenshoots-gowitness')
        for item in data:
            host = utils.get_domain(item)
            domains.append("http://" + host)
            domains.append("https://" + host)
        http_file = utils.replace_argument(
            self.options, '$WORKSPACE/screenshot/$OUTPUT-hosts.txt')
        utils.just_write(http_file, "\n".join(domains))
        utils.clean_up(http_file)
        time.sleep(2)

        # screenshots with gowitness
        cmd = "$GO_PATH/gowitness file -s $WORKSPACE/screenshot/$OUTPUT-hosts.txt -t 30 --log-level fatal --destination $WORKSPACE/screenshot/screenshoots-gowitness/ --db $WORKSPACE/screenshot/screenshoots-gowitness/gowitness.db"

        execute.send_cmd(self.options, utils.replace_argument(
            self.options, cmd), '', '', self.module_name)
        
        utils.just_waiting(self.options, self.module_name, seconds=10)

        cmd = "$GO_PATH/gowitness generate -n $WORKSPACE/screenshot/$OUTPUT-gowitness-screenshots.html  --destination  $WORKSPACE/screenshot/screenshoots-gowitness/ --db $WORKSPACE/screenshot/screenshoots-gowitness/gowitness.db"

        html_path = utils.replace_argument(
            self.options, "$WORKSPACE/portscan/$OUTPUT-gowitness-screenshots.html")
        execute.send_cmd(self.options, utils.replace_argument(
            self.options, cmd), html_path, '', self.module_name)
