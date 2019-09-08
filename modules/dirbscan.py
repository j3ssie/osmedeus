from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class DirbScan(skeleton.Skeleton):
    """docstring for DirbScan"""

    def banner(self):
        utils.print_banner("Starting DirbScan")
        utils.make_directory(self.options['WORKSPACE'] + '/directory')

