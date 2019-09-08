from modules import skeleton
from lib.core import utils


class StoScan(skeleton.Skeleton):
    """docstring for StoScan"""

    def banner(self):
        utils.print_banner("Starting Subdomain TakeOver")
        utils.make_directory(self.options['WORKSPACE'] + '/stoscan')
