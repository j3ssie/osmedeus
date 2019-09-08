from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class Probing(skeleton.Skeleton):
    """docstring for Probing"""

    def banner(self):
        utils.print_banner("Starting Probing")
        utils.make_directory(self.options['WORKSPACE'] + '/probing')
        self.options['DEBUG'] = True

    # get all subdomain previous modules
    def get_subdomains(self, command):
        utils.print_info("Joining all previous subdomain")
        final_path = command.get('requirement')
        if utils.not_empty_file(final_path):
            return
        subdomain_modules = ['SubdomainScanning',
                             'PermutationScan', 'VhostScan']
        needed_reports = []
        # get reports
        reports = report.get_report_path(self.options, module=False)
        for rep in reports:
            if rep.get('module') in subdomain_modules and 'final' in rep.get('note'):
                if utils.not_empty_file(rep.get('report_path')):
                    needed_reports.append(rep.get('report_path'))

        utils.join_files(needed_reports, final_path)
        if utils.not_empty_file(final_path):
            utils.check_output(final_path)

    # clean up for massdns result
    def clean_massdns(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        output = utils.just_read(command.get('output_path'), get_list=True)
        if not output:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False

        only_A_record, summaries, resolved = [], [], []
        for line in output:
            if '. A ' in line:
                only_A_record.append(line.split('. A ')[0])
                resolved.append(line.split('. A ')[0])
                summary = "domain|{0};;ip_address|{1}".format(
                    line.split('. A ')[0], line.split('. A ')[1])
                summaries.append(summary)
            elif '. CNAME ' in line:
                resolved.append(line.split('. CNAME ')[0])

        cleaned_output = utils.just_write(command.get(
            'cleaned_output'), "\n".join(only_A_record))

        resolved_path = utils.replace_argument(
            self.options, '$WORKSPACE/probing/resolved-$OUTPUT.txt')
        resolved_output = utils.just_write(resolved_path, "\n".join(resolved))

        if cleaned_output:
            utils.check_output(command.get('cleaned_output'))

        if resolved_output:
            utils.check_output(resolved_path)
        self.update_summaries(summaries)

    def get_domain(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        output = utils.just_read(command.get('output_path'))
        if not output:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False
        no_scheme = output.replace('https://', '').replace('http://', '')

        utils.just_write(command.get('cleaned_output'), no_scheme)
        if command.get('cleaned_output'):
            utils.check_output(command.get('cleaned_output'))

    def update_summaries(self, summaries):
        content = "\n".join(summaries)
        formatted = utils.replace_argument(
            self.options, '$WORKSPACE/probing/formatted-all-$OUTPUT.txt')
        utils.just_write(formatted, content)
        summary.push_with_file(self.options, formatted)
