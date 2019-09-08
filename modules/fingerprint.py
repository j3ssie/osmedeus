from modules import skeleton
from lib.core import utils
from lib.mode import speed
from lib.sender import execute
from lib.sender import polling
from lib.sender import report
from lib.sender import summary


class Fingerprint(skeleton.Skeleton):
    """docstring for Fingerprint"""

    def banner(self):
        utils.print_banner("Starting Fingerprint")
        utils.make_directory(self.options['WORKSPACE'] + '/fingerprint')
        utils.make_directory(self.options['WORKSPACE'] + '/fingerprint/responses')

    # clean up for massdns result
    def update_tech(self, command):
        utils.print_good('Cleaning for {0}:{1}'.format(
            command.get('banner'), command.get('post_run')))

        output = utils.just_read(command.get('output_path'), get_list=True)
        if not output:
            utils.print_bad('Output not found: {0}'.format(
                command.get('output_path')))
            return False

        tech_summaries = []
        for line in output:
            try:
                jsonl = utils.get_json(line)
                if jsonl.get('matches'):
                    domain = utils.get_domain(jsonl.get('hostname'))
                    techs = [x.get('app_name')
                            for x in jsonl.get('matches')]
                    item = "domain|{0};;technologies|{1}".format(
                        domain, ",".join(techs))
                    tech_summaries.append(item)
            except:
                pass

        cleaned_output = utils.just_write(command.get(
            'cleaned_output'), "\n".join(tech_summaries))
        if cleaned_output:
            utils.check_output(command.get('cleaned_output'))
        # update technologies to db
        summary.push_with_file(self.options, command.get('cleaned_output'))

        
