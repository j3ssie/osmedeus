import os
import sys

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
from lib.sender import send
from lib.core import utils


def login(options):
    url = options.get('remote_api') + "/auth/api/token/"
    body = {
        "username": options.get('credentials')[0],
        "password": options.get('credentials')[1]
    }
    r = send.send_post(url, body, is_json=True)
    if r.json().get('access'):
        utils.print_good("Authentication success")
        jwt = 'Osmedeus ' + r.json().get('access')
        options['JWT'] = jwt
        return options

    utils.print_bad("Authentication failed")
    return False
