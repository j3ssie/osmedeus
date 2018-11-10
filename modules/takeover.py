import os
from core import execute
from core import utils

class TakeOverScanning(object):
	def __init__(self, options):
		utils.print_banner("Scanning for Subdomain Take Over")
		self.options = options
		self.initial()

	def initial(self):
		# self.subover()
		self.subjack()

	def subover(self):
		utils.print_good('Starting SubOver')
		cmd = '$PLUGINS_PATH/SubOver/SubOver -l $WORKSPACE/subdomain/final-$TARGET.txt -v -t 100 | tee $WORKSPACE/subdomain/takeover-$TARGET-subover.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		print()

	def subjack(self):
		utils.print_good('Starting subjacj')
		cmd = '$GO_PATH/subjack -w $WORKSPACE/subdomain/final-$TARGET.txt -t 100 -timeout 30 -o $WORKSPACE/subdomain/takeover-$TARGET-subjack.txt -ssl'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		print()
