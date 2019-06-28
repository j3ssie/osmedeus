import os
from core import execute
from core import slack
from core import utils


class AssetFinding(object):
    """docstring for AssetFinding"""

    def __init__(self, options):
        utils.print_banner("AssetFinding")
        utils.make_directory(options['WORKSPACE'] + '/assets')
        self.module_name = self.__class__.__name__
        self.options = options
        self.options['CURRENT_MODULE'] = self.module_name
        self.options['SPEED'] = utils.custom_speed(self.options)
        if utils.resume(self.options, self.module_name):
            utils.print_info(
                "It's already done. use '-f' options to force rerun the module")
            return

        self.is_direct = utils.is_direct_mode(options, require_input=True)
        slack.slack_noti('status', self.options, mess={
            'title':  "{0} | {1}".format(self.options['TARGET'], self.module_name),
            'content': 'Start AssetFinding for {0}'.format(self.options['TARGET'])
        })
        self.initial()
        utils.just_waiting(self.options, self.module_name)
        slack.slack_noti('good', self.options, mess={
            'title':  "{0} | {1} ".format(self.options['TARGET'], self.module_name),
            'content': 'Done AssetFinding for {0}'.format(self.options['TARGET'])
        })
        utils.print_line()

    def initial(self):
        self.get_http()
        self.wayback_parsing()
        utils.just_waiting(self.options, self.module_name, seconds=10)
        self.get_response()
        self.linkfinder()

    # just check if http service running on it or not
    def get_http(self):
        utils.print_good('Starting httprobe')
        if self.is_direct:
            if utils.not_empty_file(self.is_direct):
                cmd = 'cat {0} | $GO_PATH/httprobe -c 100 -t 20000 -v | tee $WORKSPACE/assets/http-$OUTPUT.txt'.format(
                    self.is_direct)
            # just return if direct input is just a string
            else:
                utils.print_bad("httprobe required input as a file.")
                return None
        else:
            cmd = 'cat $WORKSPACE/subdomain/final-$OUTPUT.txt | $GO_PATH/httprobe -c 100 -t 20000 -v | tee $WORKSPACE/assets/http-$OUTPUT.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/assets/http-$OUTPUT.txt')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/assets/http-$OUTPUT.txt')
        execute.send_cmd(self.options, cmd, output_path,
                         std_path, self.module_name)
        utils.print_line()

    # grab url from waybackurl
    def wayback_parsing(self):
        utils.print_good('Starting waybackurl')
        final_domains = utils.replace_argument(
            self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')

        if self.is_direct:
            if utils.not_empty_file(self.is_direct):
                cmd = 'cat {0} | $GO_PATH/waybackurls | tee $WORKSPACE/assets/wayback-$OUTPUT.txt'.format(
                    self.is_direct)
            # just return if direct input is just a string
            else:
                cmd = 'echo {0} | $GO_PATH/waybackurls | tee $WORKSPACE/assets/wayback-$OUTPUT.txt'.format(
                    self.is_direct)
        else:

            if not utils.not_empty_file(final_domains):
                return None
            else:
                cmd = 'cat $WORKSPACE/subdomain/final-$OUTPUT.txt | $GO_PATH/waybackurls | tee $WORKSPACE/assets/wayback-$OUTPUT.txt'

        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/assets/wayback-$OUTPUT.txt')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/assets/std-wayback-$OUTPUT.std')
        execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)
        utils.print_line()

    # request for the root path to get response
    def get_response(self):
        utils.print_good('Starting meg')
        if self.is_direct:
            if utils.not_empty_file(self.is_direct):
                cmd = '$GO_PATH/meg / {0} $WORKSPACE/assets/responses/ -v -c 100'.format(
                    self.is_direct)
            # just return if direct input is just a string
            else:
                utils.print_bad("meg required input as a file.")
                return None
        else:
            cmd = '$GO_PATH/meg / $WORKSPACE/assets/http-$OUTPUT.txt $WORKSPACE/assets/responses/ -v -c 100'
        utils.make_directory(self.options['WORKSPACE'] + '/assets/responses')
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(
            self.options, '$WORKSPACE/assets/responses/index')
        std_path = utils.replace_argument(
            self.options, '$WORKSPACE/assets/responses/index')
        execute.send_cmd(self.options, cmd, output_path,
                         std_path, self.module_name)

    # finding link in http domain
    def linkfinder(self):
        utils.print_good('Starting linkfinder')

        if self.is_direct:
            if utils.not_empty_file(self.is_direct):
                http_domains = utils.just_read(self.is_direct)
            # just return if direct input is just a string
            else:
                domain = self.is_direct
                strip_domain = utils.get_domain(domain)
                if strip_domain == domain:
                    domain = 'http://' + domain
                cmd = 'python3 $PLUGINS_PATH/LinkFinder/linkfinder.py -i {0} -d -o cli | tee $WORKSPACE/assets/linkfinder/{1}-linkfinder.txt'.format(
                    domain, strip_domain)

                cmd = utils.replace_argument(self.options, cmd)
                output_path = utils.replace_argument(
                    self.options, '$WORKSPACE/assets/linkfinder/{0}-linkfinder.txt'.format(strip_domain))
                std_path = utils.replace_argument(
                    self.options, '$WORKSPACE/assets/linkfinder/{0}-linkfinder.std'.format(strip_domain))
                execute.send_cmd(self.options, cmd, output_path,
                                 std_path, self.module_name)
                return None
        else:
            if self.options['SPEED'] != 'slow':
                utils.print_good("Skipping linkfinder in quick mode")
                return None

            http_domains = utils.replace_argument(
                self.options, '$WORKSPACE/assets/http-$OUTPUT.txt')

        utils.make_directory(
            self.options['WORKSPACE'] + '/assets/linkfinder')
        if utils.not_empty_file(http_domains):
            domains = utils.just_read(http_domains)
            for domain in domains.splitlines():
                strip_domain = utils.get_domain(domain)
                cmd = 'python3 $PLUGINS_PATH/LinkFinder/linkfinder.py -i {0} -d -o cli | tee $WORKSPACE/assets/linkfinder/{1}-linkfinder.txt'.format(
                    domain, strip_domain)

                cmd = utils.replace_argument(self.options, cmd)
                output_path = utils.replace_argument(
                    self.options, '$WORKSPACE/assets/linkfinder/{0}-linkfinder.txt'.format(strip_domain))
                std_path = utils.replace_argument(
                    self.options, '$WORKSPACE/assets/linkfinder/{0}-linkfinder.std'.format(strip_domain))
                execute.send_cmd(self.options, cmd, output_path, std_path, self.module_name)

