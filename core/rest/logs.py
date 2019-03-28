import os
import json
from flask_restful import Api, Resource, reqparse
from flask import Flask, jsonify, render_template, request
import utils
current_path = os.path.dirname(os.path.realpath(__file__))

'''
get local logs command by workspace
'''

#get main json by workspace name
class Logs(Resource):
    def __init__(self, **kwargs):
        self.options = utils.reading_json(current_path + '/storages/options.json')

    def get(self, workspace):
        #
        # @TODO potential LFI here
        #
        # get specific module
        module = request.args.get('module')
        ws_name = os.path.basename(os.path.normpath(workspace))

        if ws_name in os.listdir(self.options['WORKSPACES']):

            ws_json = self.options['WORKSPACES'] + "/{0}/log.json".format(ws_name)
            if os.path.isfile(ws_json):
                raw_logs = utils.reading_json(ws_json)


                # reports[key] = utils.replace_argument(
                #     self.options, self.commands[key].get('report')).replace(self.options['WORKSPACES'], '')
                log = raw_logs
                for key in raw_logs.keys():
                    for i in range(len(raw_logs[key])):
                        log[key][i]['std_path'] = utils.replace_argument(self.options, raw_logs[key][i].get(
                            'std_path')).replace(self.options['WORKSPACES'], '')

                        log[key][i]['output_path'] = utils.replace_argument(self.options, raw_logs[key][i].get(
                            'output_path')).replace(self.options['WORKSPACES'], '')


                if module:
                    cmds = log.get(module)
                    return {'commands': cmds}
                else:
                    return log

        return 'Custom 404 here', 404

