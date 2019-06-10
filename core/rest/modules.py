import os
import json
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
from flask import Flask, jsonify, render_template, request
import utils
current_path = os.path.dirname(os.path.realpath(__file__))


'''
show report path 

'''

class Modules(Resource):
    @jwt_required
    def get(self, workspace):
        # module = request.args.get('module')
        ws_name = os.path.basename(os.path.normpath(workspace))
        options_path = current_path + \
            '/storages/{0}/options.json'.format(ws_name)
        self.options = utils.reading_json(options_path)

        if not self.options:
            return {"error": "Workspace {0} not found".format(ws_name)}

        # get commands
        commands_path = current_path + '/storages/commands.json'
        self.commands = utils.reading_json(commands_path)

        # change to current workspace instead of get from running target
        self.options['WORKSPACE'] = self.options['WORKSPACES'] + ws_name
        self.options['OUTPUT'] = ws_name

        # reports = {}
        # for key in self.commands.keys():
        #     reports[key] = []

        final_reports = []

        # reports = {}
        for key in self.commands.keys():
            final_reports.append({
                "module": key,
                "reports": []
            })

        for k in self.commands.keys():
            if "report" in self.commands[k].keys():
                report = utils.replace_argument(self.options, self.commands[k].get("report"))
                # print(report)
                if type(report) == str:
                    if utils.not_empty_file(report):
                        report_path = report.replace(
                            self.options.get('WORKSPACE'), ws_name)

                        report_item = {
                            "path": report_path,
                            "type": "html",
                        }
                        for i in range(len(final_reports)):
                            if final_reports[i].get('module') == k:
                                final_reports[i]["reports"].append(
                                    report_item)
                        # final_reports[k]["reports"].append(report_item)
                elif type(report) == list:
                    for item in report:
                        report_path = utils.replace_argument(self.options, item.get("path"))
                        if utils.not_empty_file(report_path):
                            report_path = report_path.replace(
                                self.options.get('WORKSPACE'), ws_name)

                            report_item = {
                                "path": report_path,
                                "type": item.get("type"),
                            }
                            for i in range(len(final_reports)):
                                if final_reports[i].get('module') == k:
                                    final_reports[i]["reports"].append(report_item)

        # just clean up 
        clean_reports = []
        for i in range(len(final_reports)):
            if final_reports[i].get('reports'):
                clean_reports.append(final_reports[i])

        return {'reports': clean_reports}
