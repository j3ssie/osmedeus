import os
import glob
import json
import time
from pprint import pprint
from core import execute
from core import slack
from core import utils


class Recon(object):
    """docstring for subdomain"""

    def __init__(self, options):
        utils.print_banner("Reconnaisance")
        utils.make_directory(options['WORKSPACE'] + '/recon')
        self.module_name = self.__class__.__name__
        self.options = options
        if utils.resume(self.options, self.module_name):
            utils.print_info(
                "It's already done. use '-f' options to force rerun the module")
            return
        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning Subdomain for {0}'.format(self.options['TARGET'])
        })

        self.initial()

        self.conclude()

        #this gonna run after module is done to update the main json
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done Scanning Subdomain for {0}'.format(self.options['TARGET'])
        })

    def initial(self):
        self.run()
        self.resolve_ip()
        self.technology_detection()

    def run(self):
        commands = execute.get_commands(
            self.options, self.module_name).get('routines')


        for item in commands:
            utils.print_good('Starting {0}'.format(item.get('banner')))
            #really execute it
            execute.send_cmd(self.options, item.get('cmd'), item.get(
                'output_path'), item.get('std_path'), self.module_name)
            time.sleep(1)

        utils.just_waiting(self.options, self.module_name, seconds=10, times=5)

        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))

        for item in commands:
            if "Whois" in item.get('cmd'):
                main_json["Info"]["Whois"] = {"path": item.get('output_path')}
            if "Dig" in item.get('cmd'):
                main_json["Info"]["Dig"] = {"path": item.get('output_path')}

        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)

    def technology_detection(self):
        all_subdomain_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')

        if not utils.not_empty_file(all_subdomain_path):
            return

        #add https:// prefix for all domain
        domains = utils.just_read(all_subdomain_path).splitlines()
        scheme_path = utils.replace_argument(
            self.options, '$WORKSPACE/recon/all-scheme-$OUTPUT.txt')
        utils.just_write(scheme_path, "\n".join(
            domains + [("https://" + x.strip()) for x in domains]))

        #really execute command
        cmd = '$GO_PATH/webanalyze -apps $PLUGINS_PATH/apps.json -hosts $WORKSPACE/recon/all-scheme-$OUTPUT.txt -output json -worker 20 | tee $WORKSPACE/recon/$OUTPUT-technology.json'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/recon/$OUTPUT-technology.json')
        execute.send_cmd(self.options, cmd, output_path,
                         '', self.module_name)

        utils.just_waiting(self.options, self.module_name, seconds=10, times=20)

        with open(output_path, encoding='utf-8') as o:
            data = o.read().splitlines()

        #parsing output to get technology
        techs = {}
        for line in data:
            try:
                jsonl = json.loads(line)
                if jsonl.get('matches'):
                    subdomain = jsonl.get('hostname').replace('https://', '')
                    if techs.get(subdomain):
                        techs[subdomain] += [x.get('app_name')
                                            for x in jsonl.get('matches')]
                    else:
                        techs[subdomain] = [x.get('app_name')
                                            for x in jsonl.get('matches')]
            except:
                pass
        # print(techs)

        #update the main json and rewrite that
        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))

        for i in range(len(main_json['Subdomains'])):
            sub = main_json['Subdomains'][i].get('Domain')
            if techs.get(sub):
                main_json['Subdomains'][i]["Technology"] = techs.get(sub)

        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)


    def resolve_ip(self):
        utils.print_good('Create IP for list of domain result')
        final_ip = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')

        if not utils.not_empty_file(final_ip):
            return

        cmd = '$PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -t A -o S -w $WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt $WORKSPACE/subdomain/final-$OUTPUT.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt')
        execute.send_cmd(self.options, cmd, '', '', self.module_name)

        utils.just_waiting(self.options, self.module_name, seconds=5, times=5)

        # matching IP with subdomain
        main_json = utils.reading_json(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'))

        with open(output_path, 'r') as i:
            data = i.read().splitlines()
        ips = []
        for line in data:
            if " A " in line:
                subdomain = line.split('. A ')[0]
                ip = line.split('. A ')[1]
                ips.append(ip)
                for i in range(len(main_json['Subdomains'])):
                    if subdomain == main_json['Subdomains'][i]['Domain']:
                        main_json['Subdomains'][i]['IP'] = ip

        final_ip = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')

        with open(final_ip, 'w+') as fip:
            fip.write("\n".join(str(ip) for ip in ips))

        #update the main json file
        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)


    def conclude(self):
        #just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)
        utils.print_banner("Conclusion for {0}".format(self.module_name))
