import os
import json
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
from flask import Flask, jsonify, render_template, request
import utils
current_path = os.path.dirname(os.path.realpath(__file__))


'''
Getting commands of module that have been pre-define 
'''

current_path = os.path.dirname(os.path.realpath(__file__))

class Routines(Resource):

    @jwt_required
    def get(self, workspace):
        profile = request.args.get('profile')
        module = request.args.get('module')

        ws_name = utils.get_workspace(workspace=workspace)
        
        # set default profile 
        if profile is None:
            profile = 'quick'

        routines = self.get_routine(ws_name, profile)
        if not routines:
            return {"error": "options doesn't exist for {0} workspace".format(
                workspace)}

        if module is not None:
            routines = routines.get(module)

        return {'routines': routines}
    
    # get list of commands by profile
    @jwt_required
    def get_routine(self, workspace, profile):

        # get options depend on workspace
        options_path = current_path + \
            '/storages/{0}/options.json'.format(workspace)

        commands_path = current_path + '/storages/commands.json'
        self.options = utils.reading_json(options_path)
        
        if not self.options:
            return None

        self.commands = utils.reading_json(commands_path)


        raw_routine = {}
        for key, value in self.commands.items():
            raw_routine[key] = self.commands[key].get(profile)

        routines = {}
        for module, cmds in raw_routine.items():
            routines[module] = []
            if cmds:
                for item in cmds:
                    real_item = {}
                    for k, v in item.items():
                        real_item[k] = utils.replace_argument(self.options, v)
                    routines[module].append(real_item)

        return routines

