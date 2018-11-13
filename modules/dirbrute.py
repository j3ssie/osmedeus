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
		cmd = 'dirhunt $TARGET $MORE --progress-disabled --threads 20 | tee $WORKSPACE/directory/$OUTPUT-dirhunt.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)

	def dirsearch(self):
		cmd = '$PLUGINS_PATH/dirsearch/dirsearch.py -r --plain-text-report=$WORKSPACE/directory/$OUTPUT-dirsearch.txt -u "$TARGET" -e php,asp,aspx,jsp,js -t 20 $MORE'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)


	def gobuster(self):
		cmd = '$GO_PATH/gobuster -np -t 100 -w $PLUGINS_PATH/wordlists/dir-all.txt -u $TARGET -o $WORKSPACE/directory/$OUTPUT-gobuster.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
