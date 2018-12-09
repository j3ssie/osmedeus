import os, glob
from core import execute
from core import utils

class SubdomainScanning(object):
	"""docstring for subdomain"""
	def __init__(self, options):
		utils.print_banner("Scanning Subdomain")
		utils.make_directory(options['env']['WORKSPACE'] + '/subdomain')
		self.options = options
		self.initial()

		#check if the module success or not, if not run it again
		while True:
			if os.stat(utils.replace_argument(self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')).st_size == 0:
				utils.print_bad('Something wrong with these module ... run it again')
				self.initial()
			else:
				break

	def initial(self):
		self.amass()
		self.subfinder()
		self.gobuster()
		# self.massdns()
		self.unique_result()



	def amass(self):
		utils.print_good('Starting amass')
		cmd = '$GO_PATH/amass -active -d $TARGET |tee $WORKSPACE/subdomain/$OUTPUT-amass.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/subdomain/$OUTPUT-amass.txt')

	def subfinder(self):
		utils.print_good('Starting subfinder')
		cmd = '$GO_PATH/subfinder -d $TARGET -t 100 -o $WORKSPACE/subdomain/$OUTPUT-subfinder.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/subdomain/$OUTPUT-subfinder.txt')

	def gobuster(self):
		utils.print_good('Starting gobuster')
		cmd = '$GO_PATH/gobuster -m dns -np -t 100 -w $PLUGINS_PATH/wordlists/all.txt -u $TARGET -o $WORKSPACE/subdomain/$OUTPUT-gobuster.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/subdomain/$OUTPUT-gobuster.txt')


	#don't use massdns if you use this tool via ssh cause these gonna make you lose connection
	def massdns(self):
		utils.print_good('Starting massdns')
		cmd = '$PLUGINS_PATH/massdns/scripts/subbrute.py $DOMAIN_FULL $TARGET | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o Sm -w $WORKSPACE/subdomain/raw-massdns.txt'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)

		massdns_raw = utils.replace_argument(self.options, '$WORKSPACE/subdomain/raw-massdns.txt')
		massdns_output = utils.replace_argument(self.options, '$WORKSPACE/subdomain/$OUTPUT-massdns.txt')
		if not os.path.exists(massdns_raw):
			with open(massdns_raw, 'r+') as d:
				ds = d.read().splitlines()
			for line in ds:
				newline = line.split(' ')[0][:-1]
				with open(massdns_output, 'a+') as m:
					m.write(newline + "\n")

			utils.check_output(self.options, '$WORKSPACE/subdomain/$OUTPUT-massdns.txt')


	def unique_result(self):
		utils.print_good('Unique result')
		cmd = "cat $WORKSPACE/subdomain/$OUTPUT-*.txt | sort | awk '{print tolower($0)}' | uniq >> $WORKSPACE/subdomain/final-$OUTPUT.txt"
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/subdomain/final-$OUTPUT.txt')
		

			








