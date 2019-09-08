import os
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.core import utils


def get_cmd(options):
    mode = options.get('MODE', 'general')
    url = options.get('REMOTE_API') + '/api/commands/get/?module={0}&mode={1}'.format(
        options.get('CURRENT_MODULE'), mode)

    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')
    r = send.send_get(url, data=None, headers=headers)
    if r:
        return r.json().get('commands')
    else:
        return False


# parsing command
def send_cmd(options, cmd_options):
    url = options.get('REMOTE_API') + "/api/cmd/execute/"
    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    body = cmd_options
    body['workspace'] = utils.get_workspace(options=options)
    r = send.post_without_response(url, body, headers=headers)
    # return too soon or 500 status we have something wrong
    if r and r.json().get('status') == 500:
        return False

    if not r or r.json().get('status') == 200:
        return True
