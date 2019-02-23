import time
from core import utils


from modules import initials
from modules import subdomain
from modules import takeover
from modules import screenshot
from modules import portscan
from modules import gitscan
from modules import burpstate
from modules import brutethings
from modules import dirbrute
from modules import vulnscan
from modules import cors
from modules import ipspace
from modules import sslscan
from modules import headers
from modules import conclusion

#runnning normal routine if none of module specific
def normal(options):
    #wait for flask API start
    time.sleep(2)
    utils.print_good("Running with {0} speed".format(options['speed']))

    # Create skeleton json
    initials.Initials(options)

    ##Finding subdomain
    subdomain.SubdomainScanning(options)

    ##waiting for previous module
    utils.just_waiting('SubdomainScanning')

    ##Scanning for subdomain take over
    takeover.TakeOverScanning(options)

    ##Screen shot the target on common service
    screenshot.ScreenShot(options)

    ##Scanning for CorsScan
    cors.CorsScan(options)

    # ##Discovery IP space
    ipspace.IPSpace(options)

    # ##SSL Scan
    sslscan.SSLScan(options)

    # ##Headers Scan
    headers.HeadersScan(options)

    # ##### Note: From here the module gonna take really long time for scanning service and stuff like that
    utils.print_info('This gonna take a while')

    # #Scanning all port using result from subdomain scanning and also checking vulnerable service based on version
    # portscan.PortScan(options)

    # #Starting vulnerable scan
    # vulnscan.VulnScan(options)

    # #Brute force service from port scan result
    # brutethings.BruteThings(options)

    conclusion.Conclusion(options)


def specific(options, module):
    #wait for flask API start
    time.sleep(2)

    if 'subdomain' in module:
        subdomain.SubdomainScanning(options)
        takeover.TakeOverScanning(options)
        screenshot.ScreenShot(options)
        cors.CorsScan(options)

    elif 'screenshot' in module:
        screenshot.ScreenShot(options)

    elif 'portscan' in module:
        # scanning port, service and vuln with masscan and nmap
        portscan.PortScan(options)

    elif 'vuln' in module:
        # scanning vulnerable service based on version
        vulnscan.VulnScan(options)

    elif 'git' in module:
        gitscan.GitScan(options)

    elif 'burp' in module:
        burpstate.BurpState(options)

    elif 'brute' in module or 'force' in module:
        # running brute force things based on scanning result
        brutethings.BruteThings(options)

    elif 'ip' in module:
        #Discovery IP space
        ipspace.IPSpace(options)


    elif 'headers' in module:
        #Headers Scan
        headers.HeadersScan(options)


    elif 'dir' in module:
        # run blind directory brute force directly
        dirbrute.DirBrute(options)

