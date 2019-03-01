import os, glob, json, time
from pprint import pprint
from core import execute
from core import slack
from core import utils

class SubdomainScanning(object):
    """docstring for subdomain"""
    def __init__(self, options):
        utils.print_banner("Scanning Subdomain")
        utils.make_directory(options['WORKSPACE'] + '/subdomain')
        self.module_name = self.__class__.__name__
        self.options = options

        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning Subdomain for {0}'.format(self.options['TARGET'])
        })
        self.initial()

        utils.just_waiting(self.module_name)
        #this gonna run after module is done to update the main json
        self.conclude()

        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done Scanning Subdomain for {0}'.format(self.options['TARGET'])
        })


    def initial(self):
        #just for debug purpose
        if self.options['DEBUG'] == "True":
            self.subfinder()
            self.gobuster()
        else:
            self.amass()
            self.subfinder()
            self.gobuster()
            self.massdns()



    def amass(self):
        utils.print_good('Starting amass')
        cmd = '$GO_PATH/amass -active -d $TARGET -o $WORKSPACE/subdomain/$OUTPUT-amass.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/subdomain/$OUTPUT-amass.txt')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/subdomain/std-$TARGET-amass.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)

        #log the command
        slack.slack_noti('log', self.options, mess={
            'title':  "{0} | amass | {1} | Execute".format(self.options['TARGET'], self.module_name),
            'content': '```{0}```'.format(cmd),
        })

    def subfinder(self):
        utils.print_good('Starting subfinder')
        cmd = '$GO_PATH/subfinder -d $TARGET -t 100 -o $WORKSPACE/subdomain/$OUTPUT-subfinder.txt -nW'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/subdomain/$OUTPUT-subfinder.txt')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/subdomain/std-$OUTPUT-subfinder.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)

        #log the command
        slack.slack_noti('log', self.options, mess={
            'title':  "{0} | subfinder | {1} | Execute".format(self.options['TARGET'], self.module_name),
            'content': '```{0}```'.format(cmd),
        })

    #just use massdns for directory bruteforce
    def gobuster(self):
        utils.print_good('Starting gobuster')
        
        if self.options['SPEED'] == 'slow':
            cmd = '$GO_PATH/gobuster -m dns -np -t 100 -w $PLUGINS_PATH/wordlists/all.txt -u $TARGET -o $WORKSPACE/subdomain/raw-$OUTPUT-gobuster.txt'
        elif self.options['SPEED'] == 'quick':
            cmd = '$GO_PATH/gobuster -m dns -np -t 100 -w $PLUGINS_PATH/wordlists/shorts.txt -u $TARGET -o $WORKSPACE/subdomain/raw-$OUTPUT-gobuster.txt'
        
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/raw-$OUTPUT-gobuster.txt')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/std-raw-$OUTPUT-gobuster.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)

        #log the command
        slack.slack_noti('log', self.options, mess={
            'title':  "{0} | gobuster | {1} | Execute".format(self.options['TARGET'], self.module_name),
            'content': 'Command:\n ```{0}```'.format(cmd),
        })


    def massdns(self):
        utils.print_good('Starting massdns')
        if self.options['SPEED'] == 'slow':
            cmd = '$PLUGINS_PATH/massdns/scripts/subbrute.py $DOMAIN_FULL $TARGET | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o Sm -w $WORKSPACE/subdomain/raw-massdns.txt'
        elif self.options['SPEED'] == 'quick':
            cmd = '$PLUGINS_PATH/massdns/scripts/subbrute.py $PLUGINS_PATH/wordlists/shorts.txt $TARGET | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o Sm -w $WORKSPACE/subdomain/raw-massdns.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/subdomain/raw-massdns.txt')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/subdomain/std-raw-massdns.txt')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)

        #log the command
        slack.slack_noti('log', self.options, mess={
            'title':  "{0} | massdns | {1} | Execute".format(self.options['TARGET'], self.module_name),
            'content': '```{0}```'.format(cmd),
        })

    def unique_result(self):
        #just clean up some output

        #gobuster clean up
        cmd = 'cat $WORKSPACE/subdomain/raw-$OUTPUT-gobuster.txt | cut -d ' ' -f 2 > $WORKSPACE/subdomain/$OUTPUT-gobuster.txt'
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/$OUTPUT-gobuster.txt')
        execute.send_cmd(cmd, output_path, '', self.module_name)

        #massdns clean up
        massdns_raw = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/raw-massdns.txt')
        massdns_output = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/$OUTPUT-massdns.txt')
        if not os.path.exists(massdns_raw):
            with open(massdns_raw, 'r+') as d:
                ds = d.read().splitlines()
            for line in ds:
                newline = line.split(' ')[0][:-1]
                with open(massdns_output, 'a+') as m:
                    m.write(newline + "\n")

            utils.check_output(utils.replace_argument(
                self.options, '$WORKSPACE/subdomain/$OUTPUT-massdns.txt'))

        utils.print_good('Unique result')
        cmd = "cat $WORKSPACE/subdomain/$OUTPUT-*.txt | sort | awk '{print tolower($0)}' | uniq >> $WORKSPACE/subdomain/final-$OUTPUT.txt"
        
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')
        execute.send_cmd(cmd, output_path, '', self.module_name)

        slack.slack_file(self.options, mess={
            'title':  "{0} | {1} | Output".format(self.options['TARGET'], self.module_name),
            'filename': '{0}'.format(output_path),
        })

    #update the main json file
    def conclude(self):
        self.unique_result()

        utils.print_banner("Conclusion for {0}".format(self.module_name))

        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))

        all_subdomain = utils.replace_argument(self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')
        with open(all_subdomain, 'r') as s:
            subdomains = s.read().splitlines()

        for subdomain in subdomains:
            main_json['Subdomains'].append({"Domain": subdomain})

        main_json['Modules'][self.module_name] = utils.checking_done(module=self.module_name, get_json=True)

        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)
        #write that json again
        # utils.just_write(utils.reading_json(), main_json, is_json=True)

        #logging
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(logfile)

        cmds_json = utils.checking_done(module=self.module_name, get_json=True)
        slack.slack_std(self.options, cmds_json)

        utils.print_banner("{0}".format(self.module_name))











