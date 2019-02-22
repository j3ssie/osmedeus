import os
from core import execute
from core import utils

class GitScan(object):
    """docstring for PortScan"""
    def __init__(self, options):
        utils.print_banner("Github Repo Scanning")
        utils.make_directory(options['env']['WORKSPACE'] + '/gitscan/')
        self.module_name = self.__class__.__name__
        self.options = options
        self.initial()

        self.conclude()


    def initial(self):
        self.gitleaks()
        self.truffleHog()
        self.gitrob()

    def gitleaks(self):
        cmd = '$GO_PATH/gitleaks -v --repo=$TARGET --report=$WORKSPACE/gitscan/$TARGET-gitleaks.json'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/gitscan/$TARGET-gitleaks.json')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/gitscan/std-$TARGET-gitleaks.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)

    def truffleHog(self):
        utils.print_good('Starting truffleHog')
        cmd = 'trufflehog --regex --entropy=True $TARGET | tee $WORKSPACE/gitscan/$TARGET-trufflehog.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/gitscan/$TARGET-trufflehog.txt')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/gitscan/std-$TARGET-trufflehog.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)


    def gitrob(self):
        utils.print_good('Starting gitrob')
        really_target = utils.replace_argument(self.options, '$TARGET').split('/')[3] # only get organization name

        cmd = '$GO_PATH/gitrob -save $WORKSPACE/gitscan/$TARGET-gitrob -threads 10 -github-access-token $GITHUB_API_KEY {0}'.format(really_target)
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/gitscan/$TARGET-gitrob')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/gitscan/std-$TARGET-gitrob.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)


    #update the main json file
    def conclude(self):
        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))
        main_json['Modules'][self.module_name] = checking_done(module=self.module_name, get_json=True)

        #write that json again
        utils.just_write(utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json')), main_json, is_json=True)
            
