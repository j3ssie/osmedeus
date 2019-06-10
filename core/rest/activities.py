import os
import json
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
from flask import Flask, jsonify, render_template, request
from urllib.parse import quote, unquote
from ast import literal_eval
import utils

'''
 Logging command
'''

current_path = os.path.dirname(os.path.realpath(__file__))

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

    def get_activities(self, workspace):
        ws_name = utils.get_workspace(workspace=workspace)
        activities_path = current_path + \
            '/storages/{0}/activities.json'.format(ws_name)
        self.activities = utils.reading_json(activities_path)
        if not self.activities:
            return False
        return True

    # get all activity log or by module
    @jwt_required
    def get(self, workspace):
        if not self.get_activities(workspace=workspace):
            return {"error": "activities doesn't exist for {0} workspace".format(workspace)}

        # get specific module
        module = request.args.get('module')
        if module:
            cmds = self.activities[module]
            return {'commands': cmds}
        else:
            return self.activities

    @jwt_required
    def post(self, workspace):
        if not self.get_activities(workspace=workspace):
            return {"error": "activities doesn't exist for {0} workspace".format(
                workspace)}

        data = Activities.parser.parse_args()
        cmd = data['cmd']

        module = request.args.get('module')
        # if module avalible just ignore cmd stuff
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

    # force to update activities to prevent infinity wait.
    @jwt_required
    def put(self, workspace):
        ws_name = utils.get_workspace(workspace=workspace)
        activities_path = current_path + \
            '/storages/{0}/activities.json'.format(ws_name)
        if not self.get_activities(workspace=workspace):
            return {"error": "activities doesn't exist for {0} workspace".format(
                workspace)}
        module = request.args.get('module')

        raw_activities = self.activities
        for k,v in self.activities.items():
            if k == module:
                raw_activities[k] = []

                for item in v:
                    cmd_item = item
                    cmd_item['status'] = "Done"
                    raw_activities[k].append(cmd_item)
        
        # rewrite the activities again
        utils.just_write(activities_path, raw_activities, is_json=True)

        commands = [x for x in raw_activities[module]]
        return {'commands': commands}

