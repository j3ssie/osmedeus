import os, socket
from core import execute
from core import utils
from . import vulnscan

class PortScan(object):
	"""docstring for PortScan"""
	def __init__(self, options):
		utils.print_banner("Services Scanning")
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
		domains = utils.replace_argument(self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')
		with open(domains, 'r+') as d:
			ds = d.read().splitlines()
		for domain in ds:
			try:
				ip = socket.gethostbyname(domain.strip())
				print('.',end='')
				cmd = 'echo {0} >> $WORKSPACE/subdomain/IP-$OUTPUT.txt'.format(ip)
				cmd = utils.replace_argument(self.options, cmd)
				execute.run(cmd)
			except:
				pass
		cmd = 'cat $WORKSPACE/subdomain/IP-$OUTPUT.txt | sort | uniq > $WORKSPACE/subdomain/final-IP-$OUTPUT.txt'
		cmd = utils.replace_argument(self.options, cmd)
		execute.run(cmd)

	def masscan(self):
		utils.print_good('Starting masscan')
		cmd = 'sudo masscan --rate 10000 -p0-65535 -iL $WORKSPACE/subdomain/final-IP-$OUTPUT.txt -oG $WORKSPACE/portscan/$OUTPUT-masscan.gnmap -oX $WORKSPACE/portscan/$OUTPUT-masscan.xml --wait 0'
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

	#disable because this take really long time :(
	def eyewitness_all(self):
		utils.print_good('Starting EyeWitness for all protocol')
		cmd = 'python $PLUGINS_PATH/EyeWitness/EyeWitness.py -x  $WORKSPACE/portscan/$OUTPUT-masscan.xml --web --all-protocols --prepend-https --threads 20 -d $WORKSPACE/screenshot/all/'	
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		print()
		# execute.run_as_background(cmd)

