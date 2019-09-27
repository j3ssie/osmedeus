import os
import sys

sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
from lib.sender import send
from lib.core import utils

from modules import subdomain
from modules import vhosts
from modules import permutation
from modules import probing
from modules import fingerprint
from modules import stoscan
from modules import screenshot
from modules import linkfinding
from modules import ipspace
from modules import portscan
from modules import vulnscan
from modules import dirbscan
from modules import gitscan


def handle(options):
    # input not a file just store it in default path
    # # just for debug purpose
    # print(options)
    # return
    if ',' in options.get('MODULES'):
        modules = options.get('MODULES').split(',')
    else:
        modules = [options.get('MODULES')]

    if utils.loop_grep(modules, 'sub'):
        subdomain.SubdomainScanning(options)
        # vhosts.VhostScan(options)
        # permutation.PermutationScan(options)
        probing.Probing(options)
        fingerprint.Fingerprint(options)
        # stoscan.StoScan(options)
        # screenshot.ScreenShot(options)

    if utils.loop_grep(modules, 'screen'):
        screenshot.ScreenShot(options)

    if utils.loop_grep(modules, 'fin'):
        fingerprint.Fingerprint(options)

    if utils.loop_grep(modules, 'link'):
        linkfinding.LinkFinding(options)

    if utils.loop_grep(modules, 'ip'):
        ipspace.IPSpace(options)

    if utils.loop_grep(modules, 'port'):
        portscan.PortScan(options)

    if utils.loop_grep(modules, 'vuln'):
        vulnscan.VulnScan(options)

    if utils.loop_grep(modules, 'git'):
        gitscan.GitScan(options)
    
    if utils.loop_grep(modules, 'dir'):
        dirbscan.DirbScan(options)
