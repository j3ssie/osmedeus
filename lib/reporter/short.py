import os
import sys
from tabulate import tabulate
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.sender import report
from lib.core import utils


def show(options):
    head = ['Workspaces']
    contents = report.list_workspaces(options)
    if not contents:
        utils.print_bad("No Workspace found")
        return
    print(tabulate(contents, head, tablefmt="grid"))
