import os
import sys
from tabulate import tabulate
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.sender import report
from lib.core import utils


def show(options, get_content=False):
    raw_contents = report.full_reports(options)
    # print(raw_contents)
    if not raw_contents:
        utils.print_bad("Workspace not found")
        return
    if get_content:
        reading_content(options, raw_contents)
    else:
        read_paths(raw_contents)


def reading_content(options, raw_contents):
    for element in raw_contents:
        module = element.get('module')
        reports = element.get('reports')
        # utils.print_banner(module)
        for _report in reports:

            report_path = utils.join_path(options.get(
                'WORKSPACES'), _report.get('report_path'))
            utils.print_block(report_path, tag=f'{module}:PATH')

            if _report.get('report_type') != 'html':
                # do reading file here
                utils.print_block(report_path, tag=f'{module}:READ')
                content = utils.just_read(report_path)
                print(content)
            # utils.print_line()


def read_paths(raw_contents):
    head = ['Module', 'Path']
    contents = []
    for element in raw_contents:
        module = element.get('module')
        reports = element.get('reports')
        for _report in reports:
            item = [module, _report.get('report_path')]
            contents.append(item)

        # sep = ['-'*10, '-'*30]
        # contents.append(sep)

    print(tabulate(contents, head, tablefmt="grid"))

