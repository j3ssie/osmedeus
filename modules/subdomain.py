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
		self.create_ip_result()

	# def template(self):
	# 	cmd = '$GO_PATH/amass -v -ip -d3 $WORKSPACE/$OUTPUT.html -d $TARGET'
	# 	cmd = utils.replace_argument(self.options, cmd)
	# 	# execute.run(cmd)

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

	#just for the masscan
	def create_ip_result(self):
		utils.print_good('Create IP for list of domain result')
		domains = utils.replace_argument(self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')
		with open(domains, 'r+') as d:
			ds = d.read().splitlines()
		for domain in ds:
			try:
				cmd = '''dig +short {0} |grep -oE "\b([0-9]{{1,3}}\.){{3}}[0-9]{{1,3}}\b"|head -1 >> $WORKSPACE/subdomain/IP-$OUTPUT.txt'''.format(domain)
				cmd = utils.replace_argument(self.options, cmd)
				execute.run(cmd)
			except:
				pass			








