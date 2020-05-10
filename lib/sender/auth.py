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
    try:
        if r.json().get('access'):
            utils.print_good("Authentication success")
            jwt = 'Osmedeus ' + r.json().get('access')
            options['JWT'] = jwt
            return options
    except:
        utils.print_bad("Authentication failed at: " + url)
        print('''
        [!] This might happened by running Osmedeus with sudo but the install process running with normal user
        You should install the whole Osmedeus and running it with root user.
        Or whitelist masscan + nmap in sudoers file because it's required sudo permission.
        ''')
        return False
