from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary
from lib.sender import summary


class SubdomainScanning(skeleton.Skeleton):
    """docstring for subdomain"""
    def banner(self):
        utils.print_banner("Starting Subdomain Scanning")
        utils.make_directory(self.options['WORKSPACE'] + '/subdomain')

    # clean up things and join all output path together
    def conclude(self):
        outputs = utils.get_output_path(self.commands)
        # print(outputs)
        final_output = utils.replace_argument(
            self.options, "$WORKSPACE/subdomain/final-$OUTPUT.txt")
        # print(final_output)
        outputs = utils.join_files(outputs, final_output)
        utils.check_output(final_output)
        summary.push_with_file(self.options, final_output)

    '''
    Start clean part
    '''

    # clean up gobuster
    def clean_gobuster(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))
        output = utils.just_read(command.get('output_path'))
        if not output:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False

        cleaned_output = utils.just_write(command.get(
            'cleaned_output'), output.replace('Found: ', ''))
        if cleaned_output:
            utils.check_output(command.get(
                'cleaned_output'))

    # clean up for massdns result
    def clean_massdns(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(command.get('banner'), command.get('post_run')))
        output = utils.just_read(command.get('output_path'), get_list=True)
        if not output:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False

        # only get A record 
        only_A_record = "\n".join([x.split('. A ')[0] for x in output if '. A ' in x])

        cleaned_output = utils.just_write(command.get(
            'cleaned_output'), only_A_record)
        if cleaned_output:
            utils.check_output(command.get('cleaned_output'))

    # clean up for findomain
    def clean_findomain(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(command.get('banner'), command.get('post_run')))
        output = utils.just_read(command.get('output_path'), get_list=True)
        if not output:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False

        result = []
        for line in output:
            if '>>' in line.strip():
                domain = line.strip().strip('>> ').split(' => ')[0]
                ip = line.strip().strip('>> ').split(' => ')[0]
                result.append(domain)

        cleaned_output = utils.just_write(command.get(
            'cleaned_output'), "\n".join(domain))
        if cleaned_output:
            utils.check_output(command.get('cleaned_output'))
