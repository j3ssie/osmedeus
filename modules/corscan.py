from modules import skeleton
from lib.core import utils


class CORScan(skeleton.Skeleton):
    """docstring for CORScan"""

    def banner(self):
        utils.print_banner("Scanning for CORScan")
        utils.make_directory(self.options['WORKSPACE'] + '/cors')
