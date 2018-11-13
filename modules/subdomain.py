import os, glob
from core import execute
from core import utils

class SubdomainScanning(object):
	"""docstring for subdomain"""
	def __init__(self, options):
		# super(subdomain, self).__init__()
		utils.print_banner("Scanning Subdomain")
		self.options = options
		self.initial()

	def initial(self):
		self.amass()
		self.subfinder()
		self.massdns()
		self.unique_result()

	def amass(self):
		utils.print_good('Starting amass')
		cmd = '$GO_PATH/amass -active -d $TARGET |tee $WORKSPACE/subdomain/$OUTPUT-amass.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		print()

	def subfinder(self):
		utils.print_good('Starting subfinder')
		cmd = '$GO_PATH/subfinder -d $TARGET -b -w $DOMAIN_FULL -t 100 -o $WORKSPACE/subdomain/$OUTPUT-subfinder.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		print()


	def massdns(self):
		utils.print_good('Starting massdns')
		cmd = '$PLUGINS_PATH/massdns/scripts/subbrute.py $PLUGINS_PATH/massdns/lists/names.txt $TARGET | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -t A -o S -w $WORKSPACE/subdomain/$OUTPUT-massdns.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		print()


	def unique_result(self):
		utils.print_good('Unique result')
		cmd = "cat $WORKSPACE/subdomain/$OUTPUT-*.txt | sort | awk '{print tolower($0)}' | uniq >> $WORKSPACE/subdomain/final-$OUTPUT.txt"
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)

			








