import os
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.core import utils


# parsing command
def push_with_file(options, final_output, update_type='partial'):
    utils.print_good("Update Summaries table from: {0}".format(final_output))
    ws = utils.get_workspace(options=options)
    url = options.get('REMOTE_API') + "/api/summaries/set/"
    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    body = {
        "domains_file": final_output,
        "domains": [],
        "workspace": ws,
        "update_type": update_type
    }
    # print(body)
    r = send.send_post(url, body, headers=headers, is_json=True)
    # return too soon or 500 status we have something wrong
    if r and r.json().get('status') == 200:
        return True

    return False


def get_summary(options, sum_type='full'):
    ws = utils.get_workspace(options=options)
    url = options.get('REMOTE_API') + \
        "/api/summaries/get/?workspace={0}".format(ws)
    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    r = send.send_get(url, None, headers=headers, is_json=True)
    # return too soon or 500 status we have something wrong
    if r and r.json().get('summaries'):
        return r.json().get('summaries')

    return False


def get_ip(options, field='ip'):
    ws = utils.get_workspace(options=options)
    url = options.get('REMOTE_API') + "/api/summaries/field/?workspace={0}&field={1}".format(ws, field)
    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    r = send.send_get(url, None, headers=headers, is_json=True)
    # return too soon or 500 status we have something wrong
    if r and r.json().get('summaries'):
        return r.json().get('summaries')

    return False
