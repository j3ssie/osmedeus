import os
import glob
import time

from Osmedeus.core import execute
from Osmedeus.core import slack
from Osmedeus.core import utils

class SubdomainScanning(object):
    """docstring for subdomain"""
    def __init__(self, options):
        utils.print_banner("Scanning Subdomain")
        utils.make_directory(options['WORKSPACE'] + '/subdomain')
        self.module_name = self.__class__.__name__
        self.options = options
        self.options['CURRENT_MODULE'] = self.module_name
        self.options['SPEED'] = utils.custom_speed(self.options)
        if utils.resume(self.options, self.module_name):
            utils.print_info(
                "It's already done. use '-f' options to force rerun the module")
            return
        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning Subdomain for {0}'.format(self.options['TARGET'])
        })
        self.initial()
        utils.just_waiting(self.options, self.module_name, seconds=60)
        self.conclude()

        # this gonna run after module is done to update the main json
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done Scanning Subdomain for {0}'.format(self.options['TARGET'])
        })
        utils.print_line()

    def initial(self):
        self.run()

    # grab command from commands.json
    def run(self):
        commands = execute.get_commands(self.options, self.module_name).get('routines')

        if self.options['DEBUG'] == "True":
            commands = [commands[1]]

        for item in commands:
            utils.print_good('Starting {0}'.format(item.get('banner')))
            # really execute it
            execute.send_cmd(self.options, item.get('cmd'), item.get(
                'output_path'), item.get('std_path'), self.module_name)
            time.sleep(1)
        self.massdns()
        utils.just_waiting(self.options, self.module_name, seconds=5)
        # just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)

    # running massdns for subdomain bruteforce
    # because it's really fast so it may gonna lose your internet connection
    def massdns(self):
        utils.print_good('Starting massdns')
        if self.options['SPEED'] == 'quick':
            utils.print_info('Skip massdns for quick mode')
            return None
        elif self.options['SPEED'] == 'slow':
            cmd = '$PLUGINS_PATH/massdns/scripts/subbrute.py $PLUGINS_PATH/wordlists/all.txt $TARGET | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o Sm -w $WORKSPACE/subdomain/raw-massdns.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/raw-massdns.txt')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/std-raw-massdns.txt')
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

    # finding more subdomain of subdomain :D
    def permutation(self):
        # running goaltdns
        if self.options['SPEED'] == 'quick':
            utils.print_info('Skip permutation subdomain for quick mode')
            return None
        elif self.options['SPEED'] == 'slow':
            cmd = '$GO_PATH/goaltdns -w $PLUGINS_PATH/wordlists/short-permutation.txt -l $WORKSPACE/subdomain/full-$OUTPUT.txt -o $WORKSPACE/subdomain/permutation-$OUTPUT.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/permutation-$OUTPUT.txt')
        execute.send_cmd(self.options, cmd, output_path,
                         '', self.module_name)

        utils.just_waiting(self.options, self.module_name, seconds=5)

        # resolve generated domain to verify it
        cmd2 = '$PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt --sticky -t A --verify-ip -q -o S -w $WORKSPACE/subdomain/massdns-permutation-$OUTPUT.txt $WORKSPACE/subdomain/permutation-$OUTPUT.txt'
        cmd2 = utils.replace_argument(self.options, cmd2)
        output_path2 = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt')
        execute.send_cmd(self.options, cmd, output_path2, '', self.module_name)
        utils.just_waiting(self.options, self.module_name, seconds=5)

        # strip out some false positive cause some resolvers are crazy
        raw_permutation = utils.just_read(output_path2)
        permutation_domains = []
        if raw_permutation:

            # finding false positive
            false_positive = []
            tmp_false_positive = []
            for line in raw_permutation.splitlines():
                tmp_false_positive.append(line.split('. A')[1])
            for item in tmp_false_positive:
                if tmp_false_positive.count(item) > 3:
                    false_positive.append(item)
            # really adding things
            for line in raw_permutation.splitlines():
                if line.split('. A ')[1] not in false_positive:
                    permutation_domains.append(
                        line.split('. A ')[0])

        return permutation_domains

    # just clean up some output
    def unique_result(self):
        utils.print_good('Unique result')

        # gobuster clean up
        go_raw = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/raw-$OUTPUT-gobuster.txt')
        if utils.not_empty_file(go_raw):
            go_clean = [x.split(' ')[1] for x in utils.just_read(go_raw).splitlines()]
            go_output = utils.replace_argument(
                self.options, '$WORKSPACE/subdomain/$OUTPUT-gobuster.txt')
            utils.just_write(go_output, "\n".join(go_clean))

        # massdns clean up
        massdns_raw = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/raw-massdns.txt')
        if utils.not_empty_file(massdns_raw):
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

        # joining the output
        all_output = glob.glob(utils.replace_argument(self.options,
            '$WORKSPACE/subdomain/$OUTPUT-*.txt'))
        domains = []
        for file in all_output:
            domains += utils.just_read(file).splitlines()

        output_path = utils.replace_argument(self.options, '$WORKSPACE/subdomain/full-$OUTPUT.txt')
        utils.just_write(output_path, "\n".join(set([x.strip() for x in domains])))

        # finding more subdomain
        permutation_domains = self.permutation()
        if permutation_domains:
            domains.extend(permutation_domains)
            output_path = utils.replace_argument(
                self.options, '$WORKSPACE/subdomain/full-$OUTPUT.txt')
            utils.just_write(output_path, "\n".join(
                set([x.strip() for x in domains])))
        else:
            output_path = utils.replace_argument(
                self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')
            utils.just_write(output_path, "\n".join(
                set([x.strip() for x in domains])))

        time.sleep(1)
        slack.slack_file('report', self.options, mess={
            'title':  "{0} | {1} | Output".format(self.options['TARGET'], self.module_name),
            'filename': '{0}'.format(output_path),
        })

    # update the main json file
    def conclude(self):
        self.unique_result()
        utils.print_banner("Conclusion for {0}".format(self.module_name))
        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))

        all_subdomain = utils.replace_argument(self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')

        subdomains = utils.just_read(all_subdomain).splitlines()

        for subdomain in subdomains:
            main_json['Subdomains'].append({
                "Domain": subdomain,
                "IP": "N/A",
                "Technology": ["N/A"],
                "Ports": ["N/A"],
            })

        utils.just_write(utils.replace_argument(
            self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)

        utils.print_banner("Done for {0}".format(self.module_name))
