from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class PortScan(skeleton.Skeleton):
    """docstring for PortScan"""

    def banner(self):
        utils.print_banner("Port Scanning")
        utils.make_directory(self.options['WORKSPACE'] + '/portscan')
        utils.make_directory(self.options['WORKSPACE'] + '/portscan/screenshot')
        utils.make_directory(self.options['WORKSPACE'] + '/portscan/screenshot/raw-gowitness/')
        self.delay = 500

    def update_ports(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('pre_run')))

        csv_data = utils.just_read(command.get('requirement'), get_list=True)
        if not csv_data:
            utils.print_bad('Requirement not found: {0}'.format(
                command.get('requirement')))
            return False

        result = {}
        for line in csv_data[1:]:
            host = line.split(',')[0]
            port = line.split(',')[3]
            if result.get(host, None):
                result[host] += "," + str(port).strip(',')
            else:
                result[host] = port

        # store it as format can submit to summaries
        final_result = []
        for host, ports in result.items():
            item = "ip_address|{0};;ports|{1}".format(host, ports)
            final_result.append(item)

        utils.just_write(command.get('cleaned_output'),
                         "\n".join(final_result))
        summary.push_with_file(self.options, command.get('cleaned_output'))
        # update ports to db

    def get_scheme(self, command):
        utils.print_good('Preparing for {0}:{1}'.format(
            command.get('banner'), command.get('pre_run')))

        scheme_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/scheme-$OUTPUT.txt')

        csv_data = utils.just_read(command.get('requirement'), get_list=True)

        if not csv_data:
            utils.print_bad('Requirement not found: {0}'.format(
                command.get('requirement')))
            return False
        result = []
        for line in csv_data[1:]:
            host = line.split(',')[0]
            port = line.split(',')[3]
            result.append("http://" + host + ":" + port)
            result.append("https://" + host + ":" + port)

        utils.just_write(scheme_path, "\n".join(result))
        utils.check_output(scheme_path)
        # self.update_ports(command)

    def clean_gowitness(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        real_cmd = utils.resolve_command(self.options, {
            "banner": "gowitness gen report",
            "cmd": "$GO_PATH/gowitness generate -n $WORKSPACE/portscan/screenshot/$OUTPUT-raw-gowitness.html --destination $WORKSPACE/portscan/screenshot/raw-gowitness/ --db $WORKSPACE/portscan/screenshot/gowitness.db",
            "output_path": "$WORKSPACE/portscan/screenshot/$OUTPUT-raw-gowitness.html",
        })

        execute.send_cmd(self.options, real_cmd)
        raw_html = utils.just_read(real_cmd.get('output_path'))
        if not raw_html:
            utils.print_bad('Requirement not found: {0}'.format(
                real_cmd.get('output_path')))
            return False

        local_path = utils.replace_argument(
            self.options, '$WORKSPACE/portscan/')
        real_html = raw_html.replace(local_path, '')
        utils.just_write(command.get('cleaned_output'), real_html)
        utils.check_output(command.get('cleaned_output'))
        # update screenshot in summaries
        
