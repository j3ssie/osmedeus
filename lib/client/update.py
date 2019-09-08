import os
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.core import execute

def update_():
    execute.run1(
        'git fetch --all && git reset --hard origin/master && ./install.sh')
    sys.exit(0)
