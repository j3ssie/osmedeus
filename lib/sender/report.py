import os
import sys
import time
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.core import utils


# get raw report path
def get_report_path(options, resolve=True, get_final=False, module=None):
    # ws = utils.get_workspace(options=options)
    if not module and module is not False:
        module = options.get('CURRENT_MODULE', False)

    if module:
        url = options.get(
            'REMOTE_API') + "/api/reports/raw/?module={0}".format(module)
    else:
        url = options.get('REMOTE_API') + "/api/reports/raw/"
    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    r = send.send_get(url, data=None, headers=headers)

    if not r:
        return False
    if not resolve:
        reports = r.json().get('reports')

    if resolve:
        reports = utils.resolve_commands(options, r.json().get('reports'))

    final_reports = []
    if get_final:
        for item in reports:
            if 'final' in item.get('note').lower():
                final_reports.append(item.get('report_path'))
    else:
        final_reports = reports
    if len(final_reports) == 1:
        return final_reports[0]
    return final_reports


def list_workspaces(options):
    url = options.get('REMOTE_API') + "/api/workspaces/"
    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    r = send.send_get(url, data=None, headers=headers)
    if r and r.json().get('workspaces'):
        return r.json().get('workspaces')
    return False


def full_reports(options, grouped=True):
    ws = utils.get_workspace(options=options)

    url = options.get('REMOTE_API') + \
        "/api/reports/real/?workspace={0}&grouped=true".format(ws)

    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    r = send.send_get(url, None, headers=headers, is_json=True)
    # return too soon or 500 status we have something wrong
    if r and r.json().get('reports'):
        return r.json().get('reports')
    return False
