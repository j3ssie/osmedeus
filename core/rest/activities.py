import os
import json
from flask_restful import Api, Resource, reqparse
from flask import Flask, jsonify, render_template, request
from urllib.parse import quote, unquote
from ast import literal_eval
import utils

current_path = os.path.dirname(os.path.realpath(__file__))
'''
# logging command
'''


class Activities(Resource):
    parser = reqparse.RequestParser()
    parser.add_argument('cmd',
                        type=str,
                        required=False,
                        help="This field cannot be left blank!",
                        default=None
                        )
    parser.add_argument('data',
                        type=str,
                        required=False,
                        help="This field cannot be left blank!",
                        default=None
                        )


    def __init__(self, **kwargs):
        self.activities = utils.reading_json(current_path + '/storages/activities.json')

    # get all activity log or by module
    def get(self):
        # get specific module
        module = request.args.get('module')
        if module:
            cmds = self.activities[module]
            return {'commands': cmds}
        else:
            return self.activities

    def post(self):
        data = Activities.parser.parse_args()
        cmd = data['cmd']

        module = request.args.get('module')
        #if module avalible just ignore cmd stuff
        if module:
            if cmd:
                commands = [x for x in self.activities[module]
                            if cmd in x['cmd']]
            else:
                commands = [x for x in self.activities[module]]
            return {'commands': commands}

        else:
            cmds = []
            for item in [x for x in list(self.activities.values())]:
                cmds += item
            commands = [x for x in cmds if cmd in x['cmd']]

            return {'commands': commands}
    
    def patch(self):
        data = Activities.parser.parse_args()
        raw_data = data['data']
        raw_data = unquote(data['data'])

        # print(raw_data)
        #because parser can't parse nested dict and use literal_eval to make sure we have a dict
        real_data = literal_eval(raw_data)
        module = real_data.get('module')
        content = real_data.get('content')

        activities = self.activities

        if activities.get(module) is not None:
            activities[module] += content

        utils.just_write(current_path + '/storages/activities.json', activities, is_json=True)
        return activities
