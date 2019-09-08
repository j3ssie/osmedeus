import os
import sys
from tabulate import tabulate
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.sender import report
from lib.core import utils


def show(options):
    raw_contents = report.list_workspaces(options)

    if not raw_contents:
        utils.print_bad("No Workspace found")
        return
    head = ['Available Workspaces']
    content = [[x] for x in raw_contents]
    print(tabulate(content, head, tablefmt="grid"))
