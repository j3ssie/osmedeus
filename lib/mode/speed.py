import os
import sys

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

# from lib.core import utils


def parse_speed(options):
    speed = options.get('SPEED')
    current_module = options.get('CURRENT_MODULE')
    modules = options.get('MODULES')

    quick = speed.split(';;')[0].split('|')[1:]
    raw_slow = speed.split(';;')[1].split('|')[1:]

    if raw_slow[0] == '*':
        return 'slow'

    for i in range(len(current_module)):
        if current_module[:i].lower() in raw_slow:
            return 'slow'

    return 'quick'
