from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class ScreenShot(skeleton.Skeleton):
    """docstring for ScreenShot"""

    def banner(self):
        utils.print_banner("Starting ScreenShot")
        utils.make_directory(self.options['WORKSPACE'] + '/screenshot')
        utils.make_directory(
            self.options['WORKSPACE'] + '/screenshot/raw-gowitness')

    def clean_gowitness(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        raw_html = utils.just_read(command.get('output_path'))
        if not raw_html:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False

        local_path = utils.replace_argument(
            self.options, '$WORKSPACE/screenshot/')
        real_html = raw_html.replace(local_path, '')
        utils.just_write(command.get('cleaned_output'), real_html)
        # update screenshot in summaries 
