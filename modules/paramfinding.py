from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class ParamFinding(skeleton.Skeleton):
    """docstring for ParamFinding"""

    def banner(self):
        utils.print_banner("Starting ParamFinding")
        utils.make_directory(self.options['WORKSPACE'] + '/params')
        utils.make_directory(self.options['WORKSPACE'] + '/params/raw')

