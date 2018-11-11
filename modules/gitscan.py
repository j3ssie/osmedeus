import os
from core import execute
from core import utils

class GitScan(object):
	"""docstring for PortScan"""
	def __init__(self, options):
		utils.print_banner("Github Repo Scanning")
		self.options = options
		self.initial()


	def initial(self):
		self.truffleHog()
		self.gitrob()

	def truffleHog(self):
		cmd = 'trufflehog --regex --entropy=True $TARGET | tee $WORKSPACE/gitscan/$TARGET-trufflehog.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)

	def gitrob(self):
		cmd = '$GO_PATH/gitrob -github-access-token $GITHUB_API_KEY $TARGET | tee $WORKSPACE/gitscan/$TARGET-gitrob.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)

	def gitscanner(self):
		pass
		# cmd = '$PLUGINS_PATH/reposcanner.py -r https://github.com/Dionach/reposcanner -v -a -c 30'