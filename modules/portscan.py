import os, socket
from core import execute
from core import utils
from . import vulnscan

class PortScan(object):
    """docstring for PortScan"""
    def __init__(self, options):
        utils.print_banner("Services Scanning")
        utils.make_directory(options['env']['WORKSPACE'] + '/portscan')
        self.options = options
        self.initial()

    def initial(self):
        self.create_ip_result()
        self.masscan()

        masscan_xml = utils.replace_argument(self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.xml')
        # checking output of masscan is empty or not because usually your bandwidth will not enough to scan large input 
        if utils.not_empty_file(masscan_xml):
            self.create_html()
        else:
            utils.print_bad('Masscan output empty')
            vulnscan.VulnScan(self.options)

    #just for the masscan
    def create_ip_result(self):
        utils.print_good('Create IP for list of domain result')
        cmd = '$PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -t A -o S -w $WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt $WORKSPACE/subdomain/final-$OUTPUT.txt'
        cmd = utils.replace_argument(self.options, cmd)
        execute.run(cmd)

        cmd = '''cat $WORKSPACE/subdomain/massdns-IP-$OUTPUT.txt | grep -e ' A ' |  cut -d 'A' -f 2 | tr -d ' ' > $WORKSPACE/subdomain/final-IP-$OUTPUT.txt'''
        cmd = utils.replace_argument(self.options, cmd)
        execute.run(cmd)
        utils.check_output(self.options, '$WORKSPACE/subdomain/final-IP-$OUTPUT.txt')

    def masscan(self):
        utils.print_good('Starting masscan')
        if self.options['speed'] == 'slow':
            cmd = 'sudo masscan --rate 10000 -p0-65535 -iL $WORKSPACE/subdomain/final-IP-$OUTPUT.txt -oG $WORKSPACE/portscan/$OUTPUT-masscan.gnmap -oX $WORKSPACE/portscan/$OUTPUT-masscan.xml --wait 0'
        elif self.options['speed'] == 'quick':
            utils.print_good("Only scan for single target in quick speed")
            cmd = 'sudo masscan --rate 10000 -p0-65535 $IP -oG $WORKSPACE/portscan/$OUTPUT-masscan.gnmap -oX $WORKSPACE/portscan/$OUTPUT-masscan.xml --wait 0'
        
        cmd = utils.replace_argument(self.options, cmd)
        utils.print_info("Execute: {0} ".format(cmd))
        execute.run(cmd)
        utils.check_output(self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.xml')

    def create_html(self):
        utils.print_good('Create beautify HTML report')
        cmd = 'xsltproc -o $WORKSPACE/portscan/$OUTPUT-html.html $PLUGINS_PATH/nmap-bootstrap.xsl $WORKSPACE/portscan/$OUTPUT-masscan.xml'
        cmd = utils.replace_argument(self.options, cmd)
        utils.print_info("Execute: {0} ".format(cmd))
        execute.run(cmd)
        utils.check_output(self.options, '$WORKSPACE/portscan/$OUTPUT-html.html')

    #disable because this take really long time :v
    def eyewitness_all(self):
        utils.print_good('Starting EyeWitness for all protocol')
        cmd = 'python $PLUGINS_PATH/EyeWitness/EyeWitness.py -x  $WORKSPACE/portscan/$OUTPUT-masscan.xml --web --all-protocols --prepend-https --threads 20 -d $WORKSPACE/screenshot/all/'  
        cmd = utils.replace_argument(self.options, cmd)
        utils.print_info("Execute: {0} ".format(cmd))
        print()
        # execute.run_as_background(cmd)

