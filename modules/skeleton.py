from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary
from lib.noti import slack


class Skeleton(object):
    """Base modules for all other modules inheritance"""

    def __init__(self, options):
        self.options = options
        self.options['start_time'] = utils.get_perf_time()
        self.initial()

    def banner(self):
        utils.print_block("Skeleton", tag='START')

    def initial(self):
        self.banner()
        self.module_name = self.__class__.__name__
        self.delay = 30
        self.options['CURRENT_MODULE'] = str(self.module_name)
        # check speed of the modules
        self.options['CURRENT_SPEED'] = speed.parse_speed(self.options)
        # check report file here

        if not self.resume():
            utils.print_line()
            return
        self.routine()
        self.additional_routine()
        # some noti here
        self.conclude()

    def resume(self):
        polling.clear_activities(self.options)
        # checking if final result of the module is done or not
        final_output = report.get_report_path(self.options, get_final=True)
        if utils.is_done(self.options, final_output):
            utils.print_info(
                "Module already done. Use '-f' option if you want to re run it")
            return False
        return True

    def gen_commands(self):
        self.methods = utils.get_methods(self)
        raw_commands = execute.get_cmd(self.options)
        self.pre_commands, self.mid_commands, self.post_commands = [], [], []

        if raw_commands:
            self.commands = utils.resolve_commands(self.options, raw_commands)
            for command in self.commands:
                # if command.get('pre_run') and command.get('pre_run') != '':
                #     self.pre_commands.append(command)
                if command.get('waiting') == 'last':
                    self.post_commands.append(command)
                elif command.get('waiting') == 'first':
                    self.pre_commands.append(command)
                else:
                    self.mid_commands.append(command)

    # prepare some stuff
    def routine(self):
        slack.slack_notification('status', self.options)
        self.gen_commands()
        self.really_routine(self.pre_commands)
        self.really_routine(self.mid_commands)
        self.really_routine(self.post_commands)

    def really_routine(self, commands):
        self.sub_routine(commands, kind='pre')
        self.run(commands)
        self.sub_routine(commands, kind='post')

    #  run methods in current class
    def sub_routine(self, commands, kind='post'):
        utils.print_info('Starting {0} routine for {1}'.format(
            kind, self.options.get('CURRENT_MODULE')))
        for command in commands:
            if 'pre' in kind:
                sub_method = command.get('pre_run')
            elif 'post' in kind:
                sub_method = command.get('post_run')
            if sub_method and sub_method in self.methods:
                # bypass this and get a RCE :)
                eval_string = utils.safe_eval('self.{0}(command)', sub_method)
                if eval_string:
                    eval(eval_string)
            utils.random_sleep(fixed=0.5)

    # loop through pre-defined commands and run it
    def run(self, commands):
        for command in commands:
            if command.get('cmd') == 'ignore' or command.get('cmd') == '':
                continue

            if self.options['CURRENT_SPEED'] == command.get('speed') or command.get('speed') == 'general':
                utils.print_good(
                    'Starting {0}'.format(command.get('banner')))
                if utils.check_required(command):
                    # really execute it
                    execute.send_cmd(self.options, command)
        polling.waiting(self.options, delay=self.delay)
        utils.random_sleep(fixed=0.5)

    def conclude(self):
        utils.print_line()
        utils.print_elapsed(self.options)
        utils.print_line()

    # just run additional command doesn't fit the main routine
    def additional_routine(self):
        pass
