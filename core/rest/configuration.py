import os
import json
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
from .decorators import local_only
import utils
'''
#set some config
'''

current_path = os.path.dirname(os.path.realpath(__file__))

class Configurations(Resource):
    parser = reqparse.RequestParser()
    parser.add_argument('options',
                        type=dict,
                        required=True,
                        help="This field cannot be left blank!"
                        )

    @jwt_required
    def get(self):
        # prevent reading secret from config file
        secret_things = ['USERNAME','PASSWORD', 'BOT_TOKEN', 'GITHUB_API_KEY']
        options = utils.reading_json(current_path + '/storages/options.json')
        for item in secret_things:
            del options[item]

        return options

    @local_only
    def post(self):
        current_path = os.path.dirname(os.path.realpath(__file__))
        # global options
        data = Configurations.parser.parse_args()
        options = data['options']

        utils.just_write(current_path + '/storages/options.json', options, is_json=True)

        if options.get('FORCE') == "False":
            old_log = options['WORKSPACE'] + '/log.json'
            if utils.not_empty_file(old_log) and utils.reading_json(old_log):
                utils.print_info(
                    "It's already done. use '-f' options to force rerun the module")

                raw_activities = utils.reading_json(
                    options['WORKSPACE'] + '/log.json')
                utils.just_write(current_path + '/storages/activities.json',
                                 raw_activities, is_json=True)

                return options

        
        utils.print_info("Cleasning activities log")
        #Create skeleton activities
        commands = utils.reading_json(current_path + '/storages/commands.json')
        raw_activities = {}
        for k,v in commands.items():
            raw_activities[k] = []
        utils.just_write(current_path + '/storages/activities.json',
                         raw_activities, is_json=True)

        return options
