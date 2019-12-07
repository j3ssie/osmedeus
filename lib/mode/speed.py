import os
import sys

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))


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


def excluded(options):
    current_module = options.get('CURRENT_MODULE').lower()
    exclude = options.get('EXCLUDE', '')
    if not exclude:
        return False

    if ',' in exclude:
        exclude_modules = exclude.split(',')
    else:
        exclude_modules = [exclude]

    for m in exclude_modules:
        if m == current_module:
            return True
        for i in range(len(current_module)):
            if m in current_module[:i]:
                return True

    return False
