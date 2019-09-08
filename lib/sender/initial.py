import os
import sys
from pathlib import Path

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
from lib.sender import send
from lib.core import utils


# create workspace record in the db
def init_workspace(options):
    url = options.get('remote_api') + "/api/workspace/create/"
    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    body = {
        "raw_target": options.get('raw_target'),
        'mode': options.get('mode'),
        'modules': options.get('modules', 'None'),
        'speed': options.get('speed'),
        'forced': options.get('forced'),
        'debug': options.get('debug'),
    }

    r = send.send_post(url, body, headers=headers, is_json=True)
    if r:
        options['workspace'] = r.json().get('workspace')
        # just print some log
        if r.json().get('status') == 200:
            utils.print_good("New workspace created")
        elif r.json().get('status') == 442:
            utils.print_info("Workspaces already exists. Use '-w <new workspace name>' option if you want to create new one")

        arguments = get_workspace_info(options)

        if arguments:
            options = {**options, **arguments}

            # just upper all key
            final_options = {}
            for key in options.keys():
                final_options[key.upper()] = options.get(key)

        return final_options

    utils.print_bad("Fail to create new workspace")
    return False


# get full options as well as argument ready to feed the module
def get_workspace_info(options):
    url = options.get('remote_api') + "/api/workspace/get/"
    headers = send.osmedeus_headers

    headers['Authorization'] = options.get(
        'JWT') if options.get('JWT') else options.get('jwt')

    body = {
        "workspace": options.get('workspace'),
    }

    r = send.send_post(url, body, headers=headers, is_json=True)
    if r and r.json().get('status') == 200:
        return r.json()

    utils.print_bad("Workpsace not found")
    return False
