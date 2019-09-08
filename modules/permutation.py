from modules import skeleton
from lib.core import utils


class PermutationScan(skeleton.Skeleton):
    """docstring for PermutationScan"""

    def banner(self):
        utils.print_banner("Scanning for Permutation domain")
        utils.make_directory(self.options['WORKSPACE'] + '/permutation')
