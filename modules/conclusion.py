from core import utils


class Conclusion(object):
    """Cleanning up and shutdown the flask server"""
    def __init__(self, options):
        utils.print_banner("Save log file and shutdown flask")
        # self.module_name = self.__class__.__name__
        self.options = options
        self.initial()

    def initial(self):
        self.save_log()

    def save_log(self):
        logfile = utils.replace_argument(self.options, '$WORKSPACE/log.json')
        utils.save_all_cmd(self.options, logfile)
        
        if self.options.get('CLIENT'):
            utils.just_shutdown_flask(self.options)









