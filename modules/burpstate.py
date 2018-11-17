import os
from core import execute
from core import utils

class BurpState(object):
	"""docstring for PortScan"""
	def __init__(self, options):
		self.options = options
		self.initial()

	def initial(self):
		self.linkfinder()
		self.sqlmap()
		self.sleuthql()
	
	def linkfinder(self):
		utils.print_good('Starting linkfinder')
		cmd = '$PLUGINS_PATH/linkfinder.py -i $BURPSTATE -b -o cli | tee $WORKSPACE/burp-$TARGET-linkfinder.txt'
		cmd = utils.replace_argument(self.options, cmd)
		execute.run(cmd)
		utils.print_info("Execute: {0} ".format(cmd))

	def sqlmap(self):
		utils.print_good('Starting sqlmap')
		cmd = '$PLUGINS_PATH/sqlmap/sqlmap.py -l $BURPSTATE --batch $MORE'
		cmd = utils.replace_argument(self.options, cmd)
		execute.run(cmd)
		utils.print_info("Execute: {0} ".format(cmd))

	def sleuthql(self):
		utils.print_good('Starting sleuthql')
		cmd = 'python3 $PLUGINS_PATH/sleuthql/sleuthql.py -d $TARGET -f $BURPSTATE'
		cmd = utils.replace_argument(self.options, cmd)
		execute.run(cmd)
		utils.print_info("Execute: {0} ".format(cmd))