import os
from core import execute
from core import utils

class DirBrute(object):
	"""docstring for BruteThings"""
	def __init__(self, options):
		self.options = options
		self.initial()

	def initial(self):
		self.dirhunt()
		self.dirsearch()
		self.gobuster()

	def dirhunt(self):
		utils.print_good('Starting dirhunt')
		cmd = 'dirhunt $TARGET $MORE --progress-disabled --threads 20 | tee $WORKSPACE/directory/$STRIP_TARGET-dirhunt.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)

	def dirsearch(self):
		utils.print_good('Starting dirsearch')
		cmd = '$PLUGINS_PATH/dirsearch/dirsearch.py --plain-text-report=$WORKSPACE/directory/$STRIP_TARGET-dirsearch.txt -u "$TARGET" -e php,asp,aspx,jsp,js -t 20 $MORE'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)


	def gobuster(self):
		utils.print_good('Starting gobuster')
		cmd = '$GO_PATH/gobuster -np -t 100 -w $PLUGINS_PATH/wordlists/dir-all.txt -u $TARGET -o $WORKSPACE/directory/$STRIP_TARGET-gobuster.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
