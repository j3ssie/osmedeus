import os
from core import execute
from core import utils

class CorsScan(object):
	"""docstring for PortScan"""
	def __init__(self, options):
		utils.print_banner("CORS Scanning")
		utils.make_directory(options['env']['WORKSPACE'] + '/cors/')
		self.options = options
		self.initial()


	def initial(self):
		if self.options['speed'] == 'slow':
			self.corstest()
		elif self.options['speed'] == 'quick':
			utils.print_good("Skipping for quick speed")

	def corstest(self):
		utils.print_good('Starting CORS')
		cmd = '$PLUGINS_PATH/CORStest/corstest.py -q $WORKSPACE/subdomain/final-$OUTPUT.txt | tee $WORKSPACE/cors/$TARGET-corstest.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/cors/$TARGET-corstest.txt')
