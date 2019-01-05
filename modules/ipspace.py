import os
from core import execute
from core import utils

class IPSpace(object):
	''' Scanning vulnerable service based version '''
	def __init__(self, options):
		utils.print_banner("IP Discovery")
		utils.make_directory(options['env']['WORKSPACE'] + '/ipspace/')
		self.options = options
		self.initial()

	def initial(self):
		ipOinst()

	def ipOinst(self):
		utils.print_good('Starting IPOinst')
		cmd = '$PLUGINS_PATH/IPOsint/ip-osint.py -t $TARGET -o $WORKSPACE/ipspace/$OUTPUT-ipspace.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/ipspace/$OUTPUT-ipspace.txt')
