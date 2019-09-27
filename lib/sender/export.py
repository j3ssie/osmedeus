import os
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.core import utils


def exports_to_file(options):
    ws = utils.get_workspace(options=options)
    ts = str(utils.gen_ts())
    output = options.get('OUTPUT', '')
    url = options.get('REMOTE_API') + "/api/exports/csv/"

    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')
    body = {
        "workspace": ws,
        'filename': options.get('WORKSPACE') + f'/{output}_{ts}',
    }

    r = send.send_post(url, body, headers=headers, is_json=True)
    if r and r.json().get('message'):
        return r.json().get('message')
    return False
