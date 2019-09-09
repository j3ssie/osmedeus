import os
import sys
from tabulate import tabulate
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.sender import export
from lib.core import utils


def show(options):
    content = export.exports_to_file(options)
    if not content:
        utils.print_bad("No Workspace found")
        return

    data = utils.just_read(content)
    print(data)
    utils.print_block(content, tag='OUTPUT')
