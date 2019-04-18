import os
import json
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
from flask import Flask, jsonify, render_template, request
import utils
current_path = os.path.dirname(os.path.realpath(__file__))


'''
show content of file in log command
'''

#get main json by workspace name
class Modules(Resource):

    def __init__(self, **kwargs):
        self.options = utils.reading_json(current_path + '/storages/options.json')
        self.commands = utils.reading_json(current_path + '/storages/commands.json')

    @jwt_required
    def get(self, workspace):
        module = request.args.get('module')
        ws_name = os.path.basename(os.path.normpath(workspace))
        # print(ws_name)
        # change to current workspace instead of get from running target
        self.options['WORKSPACE'] = self.options['WORKSPACES'] + ws_name
        self.options['OUTPUT'] = ws_name

        reports = {}
        for key, value in self.commands.items():
            raw_report = self.commands[key].get('report')
            reports[key] = "N/A"
            if raw_report:
                real_report = utils.replace_argument(self.options, self.commands[key].get(
                    'report'))
                if utils.not_empty_file(real_report):
                    reports[key] = real_report.replace(self.options['WORKSPACES'], '')
           

        if module is not None:
            reports = reports.get(module)

        return {'reports': reports}
