import os
import sys
from tabulate import tabulate
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.sender import summary
from lib.core import utils


def show(options):
    raw_reports = summary.get_summary(options)
    # print(raw_reports)
    if not raw_reports:
        utils.print_bad("Workspace not found")
        return

    head = ['Domain', 'IP', 'Technologies', 'Ports']
    contents = []

    for element in raw_reports:
        item = [
            element.get('domain'),
            element.get('ip_address'),
            element.get('technologies'),
            element.get('ports'),
        ]
        contents.append(item)

    print(tabulate(contents, head, tablefmt="grid"))
