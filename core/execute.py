import sys, os
import subprocess
from core import utils


def run1(command):
	os.system(command)

def run(command):
	try:
		process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)

		# Poll process for new output until finished
		while True:
			nextline = process.stdout.readline().decode('utf-8')
			# store output to log file
			if nextline == '' and process.poll() is not None:
				break
			print(nextline, end='')
			sys.stdout.flush()

		output = process.communicate()[0]
		exitCode = process.returncode

		if (exitCode == 0):
			return output
		else:
			utils.print_bad('Something went wrong with the command below: ')
			print(command)
	except:
		utils.print_bad('Something went wrong with the command below: ')
		print(command)

def run_as_background(command):
	process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
	return process