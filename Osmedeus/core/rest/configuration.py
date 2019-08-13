import os
import utils
from flask_restful import Resource, reqparse
from flask_jwt_extended import jwt_required
from configparser import ConfigParser, ExtendedInterpolation
from pathlib import Path

from .decorators import local_only

from Osmedeus.core import config
from Osmedeus.resources import *

'''
Set some config
'''

class Configurations(Resource):
    parser = reqparse.RequestParser()
    parser.add_argument('options',
                        type=dict,
                        required=True,
                        help="This field cannot be left blank!"
                        )

    # add another authen level when settings things from remote
    def verify(self, options):
        config_path = options.get('CONFIG_PATH')
        if config_path:
            # get cred from config file
            config = ConfigParser(interpolation=ExtendedInterpolation())
            config.read(config_path)
            config_username = config['Server']['username']
            config_password = config['Server']['password']

            if config_username.lower() == options.get('USERNAME').lower() \
                    and config_password.lower() == options.get('PASSWORD').lower():
                return True

        return False

    @jwt_required
    def get(self, workspace):
        ws_name = utils.get_workspace(workspace=workspace)
        options_path = config.OSMEDEUS_HOME + '/storages/{0}/options.json'.format(ws_name)
        
        # prevent reading secret from config file though API
        secret_things = ['USERNAME', 'PASSWORD', 'BOT_TOKEN', 'GITHUB_API_KEY']
        options = utils.reading_json(options_path)
        for item in secret_things:
            del options[item]

        return options

    # setting things and intitial activities log
    @local_only
    def post(self):
        # global options
        data = Configurations.parser.parse_args()
        options = data['options']

        # @TODO add another authen level when settings things from remote
        # check if credentials is the same on the config file or not
        if not self.verify(options):
            return {"error": "Can't not verify to setup config"}

        # write each workspace seprated folder
        ws_name = utils.get_workspace(options)
        utils.make_directory(config.OSMEDEUS_HOME + '/storages/{0}/'.format(ws_name))
        if not os.path.isdir(config.OSMEDEUS_HOME + '/storages/{0}/'.format(ws_name)):
            return {"error": "Can not create workspace directory with name {0} ".format(ws_name)}

        activities_path = config.OSMEDEUS_HOME + '/storages/{0}/activities.json'.format(ws_name)
        options_path = config.OSMEDEUS_HOME + '/storages/{0}/options.json'.format(ws_name)

        # consider this is settings db
        utils.just_write(options_path, options, is_json=True)

        if options.get('FORCE') == "False":
            old_log = options['WORKSPACE'] + '/log.json'
            if utils.not_empty_file(old_log) and utils.reading_json(old_log):
                utils.print_info(
                    "It's already done. use '-f' options to force rerun the module")

                raw_activities = utils.reading_json(
                    options['WORKSPACE'] + '/log.json')

                utils.just_write(activities_path,
                                 raw_activities, is_json=True)
                return options

        utils.print_info("Cleaning activities log")

        commands_path = RESOURCES_PATH.joinpath('rest/commands.json')
        commands = utils.reading_json(commands_path)

        # Create skeleton activities based on commands.json
        raw_activities = {}
        for k, v in commands.items():
            raw_activities[k] = []
        utils.just_write(activities_path,
                         raw_activities, is_json=True)

        return options
