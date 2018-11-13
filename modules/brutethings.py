import os, socket
from core import execute
from core import utils

class BruteThings(object):
	"""docstring for BruteThings"""
	def __init__(self, options):
		self.options = options
		self.routine()

	# if previous module not done this module gonna run in another routine
	def routine(self):
		require_file = utils.replace_argument(self.options, '$WORKSPACE/portscan/$OUTPUT-masscan.gnmap')
		if not os.path.exists(require_file):
			self.direct_masscan()
			self.direct_brutespray()
		else:
			self.initial()

	# normal routine
	def initial(self):
		self.brutespray()

	def brutespray(self):
		utils.print_good('Starting brutespray')
		cmd = 'python $PLUGINS_PATH/brutespray/brutespray.py --file $WORKSPACE/portscan/$OUTPUT-masscan.xml --threads 5 --hosts 5 -o $WORKSPACE/bruteforce/$OUTPUT/'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)


	#direct stuff
	def direct_masscan(self):
		utils.print_good('Starting masscan')
		ip = socket.gethostbyname(self.options['env']['STRIP_TARGET'])
		cmd = 'sudo nmap -sS -T4 -Pn -n -p- {0} -oG $WORKSPACE/portscan/$OUTPUT-nmap.gnmap -oX $WORKSPACE/portscan/$OUTPUT-nmap.xml '.format(ip)
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		print()


	def direct_brutespray(self):
		cmd = 'python $PLUGINS_PATH/brutespray/brutespray.py --file $WORKSPACE/portscan/$OUTPUT-masscan.xml --threads 5 --hosts 5 -o $WORKSPACE/bruteforce/$OUTPUT/'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)


	def patator(self):
		pass