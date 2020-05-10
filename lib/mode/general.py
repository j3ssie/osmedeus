import os
import sys
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
from lib.core import utils

from modules import subdomain
from modules import probing
from modules import formatting
from modules import fingerprint
from modules import stoscan
from modules import screenshot
from modules import linkfinding
from modules import ipspace
from modules import portscan
from modules import vulnscan
from modules import dirbscan
from modules import corscan


def handle(options):
    if utils.isFile(options.get('TARGET')):
        targets = utils.just_read(options.get('TARGET'), get_list=True)
        # loop through each target
        for target in targets:
            options['TARGET'] = target
            options['OUTPUT'] = target
            single_handle(options)
    else:
        single_handle(options)


def single_handle(options):
    subdomain.SubdomainScanning(options)
    probing.Probing(options)
    # formatting.Formatting(options)
    corscan.CORScan(options)
    fingerprint.Fingerprint(options)
    stoscan.StoScan(options)
    screenshot.ScreenShot(options)
    linkfinding.LinkFinding(options)
    ipspace.IPSpace(options)
    portscan.PortScan(options)
    # vulnscan.VulnScan(options)
    dirbscan.DirbScan(options)
