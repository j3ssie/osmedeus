import os
from core import execute
from core import utils

class BruteThings(object):
	"""docstring for PortScan"""
	def __init__(self, options):
		# super(subdomain, self).__init__()
		self.options = options
		self.initial()

	def initial(self):
		self.brutespray()


	def brutespray(self):
		cmd = 'python $PLUGINS_PATH/brutespray.py --file $WORKSPACE/portscan/$OUTPUT-masscan.gnmap --threads 5 --hosts 5'
		cmd = utils.replace_argument(self.options, cmd)
		# execute.run_as_background(cmd)
		utils.print_info("Execute: {0} ".format(cmd))

	def patator(self):
		pass