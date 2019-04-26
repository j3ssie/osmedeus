import os, time
from core import execute
from core import slack
from core import utils

class DirBrute(object):
    """docstring for DirbBrute"""
    def __init__(self, options):
        utils.print_banner("Scanning Directory")
        utils.make_directory(options['WORKSPACE'] + '/directory')
        utils.make_directory(options['WORKSPACE'] + '/directory/quick')
        utils.make_directory(options['WORKSPACE'] + '/directory/full')
        self.module_name = self.__class__.__name__
        self.options = options
        if utils.resume(self.options, self.module_name):
            utils.print_info("Detect is already done. use '-f' options to force rerun the module")
            return

        self.is_direct = utils.is_direct_mode(options, require_input=True)


        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning Directory for {0}'.format(self.options['TARGET'])
        })
        self.initial()
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start Scanning Directory for {0}'.format(self.options['TARGET'])
        })

    def initial(self):
        self.wfuzz()
        self.gobuster()

        # cmd = "$GO_PATH/gobuster -k -q -e -fw -w $PLUGINS_PATH/wordlists/quick-content-discovery.txt -o $WORKSPACE/directory/quick/{0}-gobuster.txt -t 100  -u '{0}'".format(
        # domain.strip())


    def wfuzz(self):
        utils.print_good('Starting wfuzz')
        if self.is_direct:
            domains = utils.just_read(self.is_direct).splitlines()
        else:
            #matching IP with subdomain
            main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))
            domains = [x.get('Domain') for x in main_json['Subdomains']]

        if self.options['DEBUG'] == 'True':
            domains = domains[:5]

        custom_logs = {"module": self.module_name, "content": []}

        for part in list(utils.chunks(domains, 3)):
            for domain in part:
                # cmd = "python3 $PLUGINS_PATH/dirsearch/dirsearch.py --json-report=$WORKSPACE/directory/{0}-dirsearch.json  -u '{0}' -e php,jsp,aspx,js,html -t 20 -b".format(domain.strip())
                #just strip everything to save local, it won't affect the result
                strip_domain = domain.replace(
                    'http://', '').replace('https://', '').replace('/', '-')

                cmd = "wfuzz -f $WORKSPACE/directory/quick/{1}-wfuzz.txt,raw -c -w $PLUGINS_PATH/wordlists/quick-content-discovery.txt -t 100 --sc 200,307 -u '{0}/FUZZ' | tee $WORKSPACE/directory/quick/std-{1}-wfuzz.std".format(
                    domain.strip(), strip_domain)

                cmd = utils.replace_argument(self.options, cmd)

                output_path = utils.replace_argument(
                    self.options, '$WORKSPACE/directory/quick/{0}-wfuzz.txt'.format(strip_domain))

                std_path = utils.replace_argument(
                    self.options, '$WORKSPACE/directory/quick/std-{0}-wfuzz.std'.format(strip_domain))

                execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)
                # execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name, True)
                
                # time.sleep(0.5)
                #set status to done because this gonna will be submit when all command was done
                custom_logs['content'].append({"cmd": cmd, "std_path": std_path, "output_path": output_path, "status": "Done"})
            
            #just wait couple seconds and continue but not completely stop the routine
            time.sleep(20)
        
        #submit a log
        utils.print_info('Update activities log')
        utils.update_activities(self.options, str(custom_logs))
        #just save commands
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)


    def gobuster(self):
        utils.print_good('Starting gobuster')
        if self.options['SPEED'] != 'slow':
            utils.print_good("Skipping in quick mode")
            return

        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))
        domains = [x.get('Domain') for x in main_json['Subdomains']]

        for domain in domains:
            cmd = '$GO_PATH/gobuster -k -q -e -fw -x php,jsp,aspx,html,json -w $PLUGINS_PATH/wordlists/dir-all.txt -t 100 -o $WORKSPACE/directory/$TARGET-gobuster.txt  -u "$TARGET" '

            cmd = utils.replace_argument(self.options, cmd)
            output_path = utils.replace_argument(
                self.options, '$WORKSPACE/directory/{0}-gobuster.json'.format(domain))
            std_path = utils.replace_argument(
                self.options, '$WORKSPACE/directory/std-{0}-gobuster.std'.format(domain))
            execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name, True)

    def dirhunt(self):
        utils.print_good('Starting dirhunt')
        cmd = 'dirhunt $TARGET $MORE --progress-disabled --threads 20 | tee $WORKSPACE/directory/$STRIP_TARGET-dirhunt.txt'
        cmd = utils.replace_argument(self.options, cmd)
        utils.print_info("Execute: {0} ".format(cmd))
        execute.run(cmd)

        

