import os
import json
from flask_restful import Api, Resource, reqparse

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

    def get(self):
        # print(current_path)
        options = utils.reading_json(current_path + '/storages/options.json')
        return options

    @local_only
    def post(self):
        current_path = os.path.dirname(os.path.realpath(__file__))
        print(current_path)
        # global options
        data = Configurations.parser.parse_args()
        options = data['options']
        
        utils.just_write(current_path + '/storages/options.json', options, is_json=True)
        utils.print_info("Cleasning activities log")

        #Create skeleton activities
        commands = utils.reading_json(current_path + '/storages/commands.json')
        raw_activities = {}
        for k,v in commands.items():
            raw_activities[k] = []
        utils.just_write(current_path + '/storages/activities.json',
                         raw_activities, is_json=True)

        return options
