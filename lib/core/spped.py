import os
import sys
sys.path.append(os.path.dirname(os.path.realpath(__file__)))

from core import utils


def parse_speed(options):
    speed = options.get('speed')

    quick = speed.split(';;')[0]
    slow = speed.split(';;')[0]

    print(quick, slow)
    return quick