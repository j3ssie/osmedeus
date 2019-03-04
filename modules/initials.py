import os, glob, time
from pprint import pprint
from core import execute
from core import slack
from core import utils


class Initials(object):
    """Just create skeleton and get some basic information"""
    def __init__(self, options):
        utils.print_banner("Create Skeleton JSON file")
        utils.make_directory(options['WORKSPACE'] + '/info')
        self.module_name = self.__class__.__name__
        self.options = options
        self.initial()

        # #this gonna run after module is done to update the main json
        # self.conclude()


    def initial(self):
        self.create_skeleton_json()
        self.whois()

    def create_skeleton_json(self):
        main_json = {
            "Company": utils.replace_argument(self.options, '$COMPANY'),
            "Main_domain": utils.replace_argument(self.options, '$TARGET'),
            "Whois": {"path": utils.replace_argument(self.options, '$WORKSPACE/info/$OUTPUT-whois.txt')},
            "Subdomains": [],
            "Modules": {},
            "IP Space": []
        }

        outout = utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json')

        utils.just_write(outout, main_json, is_json=True)
        utils.check_output(outout)
        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Create skeleton json'
        })

    def whois(self):
        utils.print_good('Starting Whois')
        cmd = 'whois $TARGET | tee $WORKSPACE/info/$TARGET-whois.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/info/$OUTPUT-whois.txt')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/info/std-$OUTPUT-whois.std')

        execute.send_cmd(cmd, output_path, std_path, self.module_name)
        utils.just_waiting(self.module_name, seconds=2)








