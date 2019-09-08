from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report


class VhostScan(skeleton.Skeleton):
    """docstring for VhostScan"""
    def banner(self):
        utils.print_banner("Scanning Vhosts")
        utils.make_directory(self.options['WORKSPACE'] + '/vhosts')
        utils.make_directory(self.options['WORKSPACE'] + '/vhosts/raw')

    # clean up things and join all output path together
    def conclude(self):
        outputs = report.get_output_path(self.commands)
        final_output = report.get_report_path(self.options, get_final=True)
        outputs = utils.join_files(outputs, final_output)
        utils.check_output(final_output)

    # clean up gobuster
    def clean_multi_gobuster(self, command):
        final_output = utils.replace_argument(
            self.options, "$WORKSPACE/vhosts/vhost-$OUTPUT.txt")
        # simple hack here
        raw_outputs = utils.list_files(final_output + '/../raw/', '-gobuster.txt')
        utils.join_files(raw_outputs, final_output)
        # content = final_output
        content = utils.just_read(final_output)
        if content:
            result = utils.regex_strip("\\s\\(Status.*", content)

        cleaned_output = utils.just_write(
            final_output, result.replace('Found: ', ''))
        if cleaned_output:
            utils.check_output(command.get(
                'cleaned_output'))

    # just clean up some output
    def unique_result(self):
        utils.print_good('Unique result')
        pass

