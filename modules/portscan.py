import os
from core import execute
from core import utils

class PortScan(object):
	"""docstring for PortScan"""
	def __init__(self, options):
		utils.print_banner("Services Scanning")
		self.options = options
		self.initial()

	def initial(self):
		self.aquaton()
		self.eyewitness_common()
		self.masscan()
		# self.eyewitness_all()

	def aquaton(self):
		utils.print_good('Starting aquatone')
		cmd ='cat $WORKSPACE/subdomain/final-$TARGET.txt | $GO_PATH/aquatone -out $WORKSPACE/portscan/aquatone-common/$OUTPUT'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))

	def eyewitness_common(self):
		utils.print_good('Starting EyeWitness for web')
		cmd = 'python $PLUGINS_PATH/EyeWitness.py -f $WORKSPACE/subdomain/IP-$TARGET.txt --web --prepend-https --threads 20 -d $WORKSPACE/screenshot/'	
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		print()

	def masscan(self):
		utils.print_good('Starting masscan')
		cmd = 'sudo masscan -p0-65535 --rate 1000000 -iL $WORKSPACE/subdomain/final-$TARGET.txt -oG $WORKSPACE/portscan/$OUTPUT-masscan.gnmap -oX $WORKSPACE/portscan/$OUTPUT-masscan.xml --wait 0'
		cmd = utils.replace_argument(self.options, cmd)
		# execute.run_as_background(cmd)3
		utils.print_info("Execute: {0} ".format(cmd))
		print()

	def eyewitness_all(self):
		utils.print_good('Starting EyeWitness for all protocol')
		cmd = 'python $PLUGINS_PATH/EyeWitness.py -x  $WORKSPACE/portscan/$OUTPUT-masscan.xml --web --all-protocols --prepend-https --threads 20 -d $WORKSPACE/screenshot/all/'	
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		print()
		# execute.run_as_background(cmd)

