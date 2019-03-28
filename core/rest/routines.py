import os
import json
from flask_restful import Api, Resource, reqparse
from flask import Flask, jsonify, render_template, request
import utils
current_path = os.path.dirname(os.path.realpath(__file__))


'''
Getting commands of module that have been pre-define 
'''


class Routines(Resource):
    # just return list of workspaces
    def __init__(self, **kwargs):
        self.options = utils.reading_json(current_path + '/storages/options.json')
        self.commands = utils.reading_json(current_path + '/storages/commands.json')

    def get(self):
        profile = request.args.get('profile')
        module = request.args.get('module')

        #set default profile 
        if profile is None:
            profile = 'quick'

        routines = self.get_routine(profile)

        if module is not None:
            routines = routines.get(module)

        return {'routines': routines}
    
    #get list of commands by profile
    def get_routine(self, profile):
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

