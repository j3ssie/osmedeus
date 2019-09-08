from modules import skeleton
from lib.core import utils


class GitScan(skeleton.Skeleton):
    """docstring for StoScan"""

    def banner(self):
        utils.print_banner("Starting GitScan")
        utils.make_directory(self.options['WORKSPACE'] + '/gitscan')
