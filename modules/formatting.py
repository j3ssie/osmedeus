from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class Formatting(skeleton.Skeleton):
    """docstring for Formatting"""

    def banner(self):
        utils.print_banner("Start Formatting")
        utils.make_directory(self.options['WORKSPACE'] + '/formatted')

    # just disable slack for this module
    def additional_routine(self):
        pass

    # clean up for massdns result
    def clean_massdns(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        output = utils.just_read(command.get('output_path'), get_list=True)
        if output:
            only_A_record, resolved, ips = [], [], []
            for line in output:
                if '. A ' in line:
                    ip = line.split('. A ')[1].strip()
                    domain = line.split('. A ')[0]
                    only_A_record.append(domain)
                    ips.append(ip)
                    resolved.append(line.split('. A ')[0])

            cleaned_output = utils.just_write(command.get(
                'cleaned_output'), "\n".join(ips))

            if cleaned_output:
                utils.check_output(command.get('cleaned_output'))

        self.join_ip(command)

    def join_ip(self, command):
        cleaned_output = utils.just_read(command.get('cleaned_output'), get_list=True)
        raw_input = utils.just_read(
            command.get('requirement'), get_list=True)

        result = []
        for line in raw_input:
            if utils.valid_ip(line.strip()):
                result.append(line)
        if cleaned_output:
            result = list(set(result + cleaned_output))
        else:
            result = list(set(result))

        if result:
            utils.just_write(command.get('cleaned_output'), "\n".join(result))
            summaries = []
            for item in result:
                summary = f"domain|{item};;ip_address|{item}"
                summaries.append(summary)
            self.update_summaries(summaries)

    def update_summaries(self, summaries):
        content = "\n".join(summaries)
        formatted = utils.replace_argument(
            self.options, '$WORKSPACE/formatted/formatted-$OUTPUT.txt')
        utils.just_write(formatted, content)
        summary.push_with_file(self.options, formatted)
