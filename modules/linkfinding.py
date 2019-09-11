from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class LinkFinding(skeleton.Skeleton):
    """docstring for LinkFinding"""

    def banner(self):
        utils.print_banner("Starting Linkfinding")
        utils.make_directory(self.options['WORKSPACE'] + '/links')
        utils.make_directory(self.options['WORKSPACE'] + '/links/raw')

    def clean_waybackurls(self, command):
        raw_output = command.get('output_path')
        final_output = command.get('cleaned_output')
        utils.strip_blank_line(final_output, raw_output)

    def clean_linkfinder(self, command):
        final_output = command.get('cleaned_output')
        # simple hack here
        raw_outputs = utils.list_files(final_output + '/../raw/', '.txt')
        utils.join_files(raw_outputs, final_output)
        utils.check_output(final_output)
        # update screenshot in summaries
