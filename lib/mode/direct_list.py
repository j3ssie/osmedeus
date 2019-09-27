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
from modules import formatting
from modules import dirbscan
from modules import gitscan
from modules import corscan


# change mode to direct
def single_handle(options, modules):
    options['MODE'] = 'direct'
    if utils.isFile(options.get('TARGET')):
        targets = utils.just_read(options.get('TARGET'), get_list=True)

    for target in targets:
        options['TARGET'] = target
        # run each line as a direct mode
        if utils.loop_grep(modules, 'sub'):
            subdomain.SubdomainScanning(options)
            probing.Probing(options)
        if utils.loop_grep(modules, 'screen'):
            screenshot.ScreenShot(options)
        if utils.loop_grep(modules, 'take'):
            stoscan.StoScan(options)


def handle(options):
    # just for debug purpose
    # print(options)
    if ',' in options.get('MODULES'):
        modules = options.get('MODULES').split(',')
    else:
        modules = [options.get('MODULES')]

    # run each line as a direct mode
    if utils.loop_grep(modules, 'sub'):
        single_handle(options, modules)
        return

    formatting.Formatting(options)

    # support direct list natively
    if utils.loop_grep(modules, 'screen'):
        screenshot.ScreenShot(options)

    if utils.loop_grep(modules, 'takeover'):
        stoscan.StoScan(options)

    if utils.loop_grep(modules, 'fingerprint'):
        fingerprint.Fingerprint(options)

    if utils.loop_grep(modules, 'port'):
        portscan.PortScan(options)

    if utils.loop_grep(modules, 'vuln'):
        vulnscan.VulnScan(options)

    if utils.loop_grep(modules, 'git'):
        gitscan.GitScan(options)

    if utils.loop_grep(modules, 'dir'):
        dirbscan.DirbScan(options)

    if utils.loop_grep(modules, 'cors'):
        corscan.CORScan(options)

    if utils.loop_grep(modules, 'link'):
        linkfinding.LinkFinding(options)

    if utils.loop_grep(modules, 'ip'):
        ipspace.IPSpace(options)

    # @TODO add headers, ssl, burp
