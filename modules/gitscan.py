import os
from core import execute
from core import slack
from core import utils

class GitScan(object):
    """docstring for PortScan"""
    def __init__(self, options):
        utils.print_banner("Github Repo Scanning")
        utils.make_directory(options['WORKSPACE'] + '/gitscan/')
        self.module_name = self.__class__.__name__
        self.options = options

        if utils.resume(self.options, self.module_name):
            utils.print_info(
                "It's already done. use '-f' options to force rerun the module")
            return
        self.is_direct = utils.is_direct_mode(options, require_input=False)


        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Github Repo Scanning for {0}'.format(self.options['TARGET'])
        })
        self.initial()

        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Github Repo Scanning for {0}'.format(self.options['TARGET'])
        })


    def initial(self):
        self.run()
        self.gitrob()

    def run(self):
        commands = execute.get_commands(self.options, self.module_name).get('routines')
        for item in commands:
            utils.print_good('Starting {0}'.format(item.get('banner')))
            #really execute it
            execute.send_cmd(self.options, item.get('cmd'), item.get(
                'output_path'), item.get('std_path'), self.module_name)

        utils.just_waiting(self.options, self.module_name, seconds=2)
        #just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)

    def gitrob(self):
        utils.print_good('Starting gitrob')
        really_target = utils.replace_argument(self.options, '$TARGET').split('/')[3] # only get organization name

        cmd = '$GO_PATH/gitrob -save $WORKSPACE/gitscan/$TARGET-gitrob -threads 10 -github-access-token $GITHUB_API_KEY {0}'.format(really_target)
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/gitscan/$TARGET-gitrob')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/gitscan/std-$TARGET-gitrob.std')
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

