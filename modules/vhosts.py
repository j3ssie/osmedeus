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

    # clean up gobuster
    def clean_vhosts_gobuster(self, command):
        final_output = utils.replace_argument(
            self.options, "$WORKSPACE/vhosts/vhosts-$OUTPUT.txt")
        raw_outputs = utils.replace_argument(
            self.options, "$WORKSPACE/vhosts/raw-summary-$OUTPUT.txt")

        content = utils.just_read(raw_outputs)
        if not content:
            return

        result = utils.regex_strip("\\s\\(Status.*", content)
        cleaned_output = utils.just_write(
            final_output, result.replace('Found: ', ''))
        if cleaned_output:
            utils.check_output(command.get(
                'cleaned_output'))
