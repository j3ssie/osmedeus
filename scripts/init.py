#!/usr/bin/env python3
# -*- coding: utf-8 -*-

'''
This script gonna generate init config and data
'''

import os
import sys
import time
import hashlib
import argparse
from pathlib import Path
import django

# Load default path of extensions folder
# WARNING: change this if you custom something
BASE_DIR = Path(os.path.dirname(
    os.path.dirname(os.path.abspath(__file__))))
sys.path.append(str(BASE_DIR.joinpath('server')))

# loading django setup
os.environ.setdefault("DJANGO_SETTINGS_MODULE", "rest.settings")
django.setup()

from api.models import *

from core import common
from core import dbutils
from core import utils

from django.contrib.auth.models import User


def config_client(username, password, remote, config_path='~/.osmedeus/client.conf'):
    # checking for config path
    config_path = utils.absolute_path(config_path)
    if os.path.isfile(config_path):
        utils.print_info('Loading config file from: {0}'.format(config_path))
    else:
        utils.print_info('New config file created: {0}'.format(config_path))
        utils.file_copy(utils.TEMPLATE_CLIENT_CONFIG, config_path)

    configs = utils.just_read_config(config_path, raw=True)

    # write the config again
    configs.set('Server', 'remote_api', remote)
    configs.set('Server', 'username', username)
    configs.set('Server', 'password', password)
    with open(config_path, 'w+') as configfile:
        configs.write(configfile)

    return True


def create_user(username, password):
    try:
        user = User.objects.create_user(username, password=password)
        user.is_superuser = True
        user.is_staff = True
        user.save()
        utils.print_good(f'{username} user created with password {password}')
    except:
        user = User.objects.get(username=username)
        user.set_password(password)
        user.save()
        utils.print_info("{0} user already exist".format(username))


def main(args):
    random = hashlib.md5(str(int(time.time())).encode()).hexdigest()[:8]

    remote = args.remote if args.remote else 'http://127.0.0.1:8000'
    username = args.username if args.username else 'osmedeus'
    password = args.password if args.password else str(random)
    create_user(username, password)

    # load command from default workflow 
    dbutils.gen_default_config('~/.osmedeus/server.conf')
    dbutils.internal_parse_commands()
    # create server config file
    dbutils.load_default_config()
    config_client(username, password, remote)


parser = argparse.ArgumentParser(description="Initial setup")
parser.add_argument('-u', '--username', action='store', dest='username', help='username')
parser.add_argument('-p', '--password', action='store', dest='password', help='password')
parser.add_argument('-r', '--remote', action='store', dest='remote', help='remote')
args = parser.parse_args()
main(args)
