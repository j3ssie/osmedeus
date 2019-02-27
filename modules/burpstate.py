import os
from core import execute
from core import slack
from core import utils

class BurpState(object):
    """docstring for PortScan"""
    def __init__(self, options):
        utils.print_banner("Scanning through BurpState")
        utils.make_directory(options['WORKSPACE'] + '/burpstate/')
        self.module_name = self.__class__.__name__
        self.options = options
        slack.slack_info(self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning BurpState for {0}'.format(self.options['TARGET'])
        })
        self.initial()
        slack.slack_good(self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning BurpState for {0}'.format(self.options['TARGET'])
        })
        self.conclude()

    def initial(self):
        self.linkfinder()
        self.sqlmap()
        self.sleuthql()
    
    def linkfinder(self):
        utils.print_good('Starting linkfinder')
        cmd = '$PLUGINS_PATH/linkfinder.py -i $BURPSTATE -b -o cli | tee $WORKSPACE/burp-$TARGET-linkfinder.txt'
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/burp-$TARGET-linkfinder.txt')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/burp-$TARGET-linkfinder.txt')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)


    def sqlmap(self):
        utils.print_good('Starting sqlmap')
        cmd = '$PLUGINS_PATH/sqlmap/sqlmap.py -l $BURPSTATE --batch $MORE'
        cmd = utils.replace_argument(self.options, cmd)
        execute.send_cmd(cmd, '', '', self.module_name)


    def sleuthql(self):
        utils.print_good('Starting sleuthql')
        cmd = 'python3 $PLUGINS_PATH/sleuthql/sleuthql.py -d $TARGET -f $BURPSTATE'
        cmd = utils.replace_argument(self.options, cmd)
        execute.send_cmd(cmd, '', '', self.module_name)


    #update the main json file
    def conclude(self):
        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))
        main_json['Modules'][self.module_name] = utils.checking_done(module=self.module_name, get_json=True)

        #write that json again
        utils.just_write(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)
