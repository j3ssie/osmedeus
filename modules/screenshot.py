import os, socket
from core import execute
from core import utils

class ScreenShot(object):
	"""Screenshot all domain on common service"""
	def __init__(self, options):
		utils.print_banner("Services Scanning")
		self.options = options
		self.initial()

	def initial(self):
		self.aquaton()
		self.eyewitness_common()

	def aquaton(self):
		utils.print_good('Starting aquatone')
		cmd ='cat $WORKSPACE/subdomain/final-$TARGET.txt | $GO_PATH/aquatone -threads 20 -out $WORKSPACE/subdomain/$OUTPUT-aquatone.html'
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/subdomain/$OUTPUT-aquatone.html')

	def eyewitness_common(self):
		utils.print_good('Starting EyeWitness for web')
		cmd = 'python $PLUGINS_PATH/EyeWitness/EyeWitness.py -f $WORKSPACE/subdomain/final-$TARGET.txt --web --prepend-https --threads 20 -d $WORKSPACE/screenshot/'	
		cmd = utils.replace_argument(self.options, cmd)
		utils.print_info("Execute: {0} ".format(cmd))
		execute.run(cmd)
		utils.check_output(self.options, '$WORKSPACE/screenshot/')
		
