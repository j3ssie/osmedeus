import os
import sys
import shutil
import difflib

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.core import utils
from lib.noti import slack_noti

'''
Only avaliable in pro version
'''


def check_diff(options, reports):
    pass


def push_to_db(options, noti_options):
    pass


def parse_diff(options, result, old_path, new_path):
    pass


def diff_content(old_path, new_path):
    pass
