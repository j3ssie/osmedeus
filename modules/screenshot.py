import os, time
from core import execute
from core import utils

class ScreenShot(object):
    """Screenshot all domain on common service"""
    def __init__(self, options):
        utils.print_banner("ScreenShot the target")
        utils.make_directory(options['env']['WORKSPACE'] + '/screenshot')
        self.module_name = self.__class__.__name__
        self.options = options

        self.initial()
        #check if the screenshot success or not, if not run it again
        # while True:
        #     if not os.listdir(utils.replace_argument(self.options, '$WORKSPACE/screenshot/')):
        #         utils.print_bad('Something wrong with these module ... run it again')
        #         self.initial()
        #         utils.just_waiting(self.module_name)
        #     else:
        #         break




    def initial(self):
        self.aquaton()
        # really slow the flow so disable for now
        # self.eyewitness_common()
        utils.just_waiting(self.module_name, seconds=10)
        #this gonna run after module is done to update the main json
        self.conclude()

    def aquaton(self):
        utils.print_good('Starting aquatone')
        cmd ='cat $WORKSPACE/subdomain/final-$TARGET.txt | $GO_PATH/aquatone -threads 20 -out $WORKSPACE/screenshot/$OUTPUT-aquatone.html'
        
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/screenshot/$OUTPUT-aquatone.html')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/screenshot/std-$OUTPUT-aquatone.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)

    def eyewitness_common(self):
        utils.print_good('Starting EyeWitness for web')
        cmd = 'python $PLUGINS_PATH/EyeWitness/EyeWitness.py -f $WORKSPACE/subdomain/final-$TARGET.txt --web --prepend-https --threads 20 -d $WORKSPACE/screenshot/eyewitness-$TARGET/' 
        cmd = utils.replace_argument(self.options, cmd)
        output_path = utils.replace_argument(self.options, '$WORKSPACE/screenshot/')
        std_path = utils.replace_argument(self.options, '$WORKSPACE/screenshot/std-eyewitness-$TARGET.std')
        execute.send_cmd(cmd, output_path, std_path, self.module_name)

        

    #update the main json file
    def conclude(self):
        main_json = utils.reading_json(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'))
        main_json['Modules'][self.module_name] = utils.checking_done(module=self.module_name, get_json=True)

        #write that json again
        utils.just_write(utils.replace_argument(self.options, '$WORKSPACE/$COMPANY.json'), main_json, is_json=True)
        
        #logging
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(logfile)

        utils.print_banner("{0} Done".format(self.module_name))

