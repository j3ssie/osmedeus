import os
import sys
import time
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.core import utils


def waiting(options, delay=20, times=0):
    elapsed_time = 0
    if times:
        count = 0
    module_name = options.get('CURRENT_MODULE', False)
    utils.print_info('Waiting for {0} module'.format(module_name))
    checking = poll_status(options)

    while checking:
        time.sleep(delay)
        if not times:
            # just don't print this too much
            if ((elapsed_time / delay) % 10) == 0:
                utils.print_info('Waiting for {0} module'.format(module_name))
                time.sleep(delay)
        if times:
            utils.print_info('Waiting for {0} module {1}/{2}'.format(module_name, str(count), str(times)))
            if count == int(times):
                poll_status(options, forced=True)
                utils.print_bad("Something bad with {0} module but force to continue".format(module_name))
                break
            count += 1
        checking = poll_status(options)
        # print(checking)


# clear old stuff because it's may be failed
# continious will check output file so don't worry
def clear_activities(options):
    ws = utils.get_workspace(options=options)
    module = options.get('CURRENT_MODULE', False)

    url = options.get('REMOTE_API') + "/api/activities/clear/"

    body = {
        "workspace": ws,
        "module": module,
    }

    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')
    r = send.send_post(url, body, headers=headers, is_json=True)
    if r and r.json().get('status') == 200:
        utils.print_good("Clean old activities for {0}:{1}".format(ws, module))
        return True

    return False


# parsing command
def poll_status(options, forced=False):
    # /api/activities/get/?workspace=duckduckgo.com&modules=SubdomainScanning
    ws = utils.get_workspace(options=options)
    module = options.get('CURRENT_MODULE', False)
    if module:
        url = options.get('REMOTE_API') + "/api/activities/get/?workspace={0}&module={1}".format(ws, module)
    else:
        url = options.get('REMOTE_API') + "/api/activities/get/?workspace={0}".format(ws)

    headers = send.osmedeus_headers
    headers['Authorization'] = options.get('JWT')

    if forced:
        r = send.send_post(url, data=None, headers=headers)
    else:
        r = send.send_get(url, data=None, headers=headers)

    if r and r.json().get('status') == 'Done':
        return False

    if not r or r.json().get('status') != 'Done':
        return True
