import os
from core import execute
from core import utils

class VulnScan(object):
	''' Scanning vulnerable service based version '''
	def __init__(self, options):
		utils.print_banner("Vulnerable Scanning")
		utils.make_directory(options['env']['WORKSPACE'] + '/vulnscan/')
		self.options = options
		self.initial()

	def initial(self):
		#check if this module is done from portscan or not
		require_file = utils.replace_argument(self.options, '$WORKSPACE/vulnscan/$OUTPUT-nmap.xml')
		if not os.path.exists(require_file):
			self.nmap_vuln()
			self.create_html()
		else:
			utils.print_good('This module is already run')

	def nmap_vuln(self):
		utils.print_good('Starting nmap vulnerable scan')
		cmd = 'nmap -T4 -Pn -n -sSV -p- $STRIP_TARGET --script vulners --oA $WORKSPACE/vulnscan/$OUTPUT-nmap'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/vulnscan/$TARGET-nmap.xml')

	def create_html(self):
		utils.print_good('Create beautify HTML report')
		cmd = 'xsltproc -o $WORKSPACE/vulnscan/$OUTPUT.html $PLUGINS_PATH/nmap-bootstrap.xsl $WORKSPACE/vulnscan/$OUTPUT-nmap.xml'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/vulnscan/$TARGET.html')
