from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class VulnScan(skeleton.Skeleton):
    """docstring for VulnScan"""

    def banner(self):
        utils.print_banner("Vulnerabily Scanning")
        utils.make_directory(self.options['WORKSPACE'] + '/vulnscan')
        utils.make_directory(
            self.options['WORKSPACE'] + '/vulnscan/details')
        utils.make_directory(
            self.options['WORKSPACE'] + '/vulnscan/report')
        utils.make_directory(
            self.options['WORKSPACE'] + '/vulnscan/screenshot')
        utils.make_directory(
            self.options['WORKSPACE'] + '/vulnscan/screenshot/raw-gowitness/')
        self.delay = 1200

    # only run on direct mode
    def get_domains(self, command):
        pass

    def gen_summary(self, command):
        summary_path = utils.replace_argument(
            self.options, '$WORKSPACE/vulnscan/summary-$OUTPUT.csv')
        sum_head = 'IP,Host,OS,Proto,Port,Service,Product,Service FP,NSE Script ID,NSE Script Output,Notes'

        details_folder = utils.replace_argument(
            self.options, '$WORKSPACE/vulnscan/details/')
        details = utils.list_files(details_folder, '.csv')
        summary_data = [sum_head]
        for detail in details:
            really_detail = utils.just_read(detail, get_list=True)
            if really_detail:
                summary_data.append("\n".join(really_detail[1:]))

        utils.just_write(summary_path, "\n".join(summary_data))

    # update ports found in db
    def update_ports(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        csv_data = utils.just_read(command.get('output_path'), get_list=True)
        if not csv_data:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False

        result = {}
        for line in csv_data[1:]:
            host = line.split(',')[0]
            port = line.split(',')[3]
            if result.get('host', None):
                result[host] += "," + str(port).strip(',')
            else:
                result[host] = port

        final_result = []
        for host, ports in result.items():
            item = "ip_address|{0};;ports|{1}".format(host, ports)
            final_result.append(item)

        utils.just_write(command.get('cleaned_output'),
                         "\n".join(final_result))
        # update ports to db

    # get all scheme from csv summary
    def get_scheme(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        csv_data = utils.just_read(command.get('requirement'), get_list=True)
        if not csv_data:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False
        summaries, result = [], []
        for line in csv_data[1:]:
            # print(line)
            if ',' not in line or len(line.split(',')) < 4:
                continue
            host = line.split(',')[0]
            port = line.split(',')[4]
            service = line.split(',')[5] + "/" + line.split(',')[6]
            result.append("http://" + host + ":" + port)
            result.append("https://" + host + ":" + port)
            sum_line = f"domain|{host};;ip_address|{host};;ports|{port};;technologies|{service}"
            summaries.append(sum_line)
            # print(sum_line)

        scheme_path = utils.replace_argument(
            self.options, '$WORKSPACE/vulnscan/scheme-$OUTPUT.txt')
        utils.just_write(scheme_path, "\n".join(result))

        # update summaries table
        formatted_summary = utils.replace_argument(
            self.options, '$WORKSPACE/vulnscan/formatted-summary-$OUTPUT.txt')
        utils.just_write(formatted_summary, "\n".join(summaries))
        summary.push_with_file(self.options, formatted_summary)

    def clean_gowitness(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        command = utils.resolve_command(self.options, {
            "banner": "gowitness gen report",
            "cmd": "$GO_PATH/gowitness generate -n $WORKSPACE/vulnscan/screenshot/$OUTPUT-raw-gowitness.html --destination $WORKSPACE/vulnscan/screenshot/raw-gowitness/ --db $WORKSPACE/vulnscan/screenshot/gowitness.db",
            "output_path": "$WORKSPACE/vulnscan/screenshot/$OUTPUT-raw-gowitness.html",
        })
        execute.send_cmd(self.options, command)
        raw_html = utils.just_read(command.get('output_path'))
        if not raw_html:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False

        local_path = utils.replace_argument(
            self.options, '$WORKSPACE/vulnscan/screenshot/')
        real_html = raw_html.replace(local_path, '/')
        utils.just_write(command.get('cleaned_output'), real_html)
        # update screenshot in summaries
