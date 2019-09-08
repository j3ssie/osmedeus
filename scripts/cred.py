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


from core import utils
from django.contrib.auth.models import User


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
        utils.print_info(f"{username} user already exist")


def main(args):
    random = hashlib.md5(str(int(time.time())).encode()).hexdigest()[:8]
    username = args.username if args.username else 'osmedeus'
    password = args.password if args.password else str(random)
    create_user(username, password)


parser = argparse.ArgumentParser(description="User mangement setup")
parser.add_argument('-u', '--username', action='store', dest='username', help='username')
parser.add_argument('-p', '--password', action='store', dest='password', help='password')
args = parser.parse_args()
main(args)
