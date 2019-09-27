import os
import sys
import shutil


sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.sender import send
from lib.core import utils
from lib.monitor import compare

'''
Only avaliable in pro version
'''


# return compare_path to options
def init_backup(options):
    pass


# only keep 3 newest result
def clean_oldbackup(options):
    pass
