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


# change mode to direct
def single_handle(options, modules):
    options['MODE'] = 'direct'
    if not utils.isFile(options.get('TARGET')):
        targets = utils.just_read(options.get('TARGET'), get_list=True)

    for target in targets:
        options['TARGET'] = target
        # run each line as a direct mode
        if utils.loop_grep(modules, 'subdomain'):
            subdomain.SubdomainScanning(options)
            vhosts.VhostScan(options)
            permutation.PermutationScan(options)
            probing.Probing(options)
            screenshot.ScreenShot(options)
            stoscan.StoScan(options)
            fingerprint.Fingerprint(options)

        if utils.loop_grep(modules, 'link'):
            linkfinding.LinkFinding(options)

        if utils.loop_grep(modules, 'ip'):
            ipspace.IPSpace(options)


def handle(options):
    # input not a file just store it in default path
    # just for debug purpose
    # print(options)
    # return

    if ',' in options.get('MODULES'):
        modules = options.get('MODULES').split(',')
    else:
        modules = [options.get('MODULES')]

    formatting.Formatting(options)
    
    # return 
    # run each line as a direct mode
    if utils.loop_grep(modules, 'subdomain'):
        single_handle(options, modules)
    if utils.loop_grep(modules, 'link'):
        single_handle(options, modules)
    if utils.loop_grep(modules, 'ip'):
        single_handle(options, modules)

    # support direct list natively
    if utils.loop_grep(modules, 'screen'):
        screenshot.ScreenShot(options)

    if utils.loop_grep(modules, 'takeover'):
        stoscan.StoScan(options)

    if utils.loop_grep(modules, 'fin'):
        fingerprint.Fingerprint(options)

    if utils.loop_grep(modules, 'port'):
        portscan.PortScan(options)

    if utils.loop_grep(modules, 'vuln'):
        vulnscan.VulnScan(options)

    if utils.loop_grep(modules, 'git'):
        gitscan.GitScan(options)

    # @TODO cors, headers, ssl, burp
