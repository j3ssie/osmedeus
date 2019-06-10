import os
import json
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
from flask import Flask, jsonify, render_template, request
import utils
current_path = os.path.dirname(os.path.realpath(__file__))

'''
get local logs command by workspace
'''

class Logs(Resource):
    # def __init__(self, **kwargs):
    #     self.options = utils.reading_json(
    #         current_path + '/storages/options.json')

    @jwt_required
    def get(self, workspace):
        # get options depend on workspace
        ws_name = utils.get_workspace(workspace=workspace)
        options_path = current_path + \
            '/storages/{0}/options.json'.format(ws_name)
        self.options = utils.reading_json(options_path)

        module = request.args.get('module')
        ws_name = os.path.basename(os.path.normpath(workspace))

        if ws_name in os.listdir(self.options['WORKSPACES']):
            ws_json = self.options['WORKSPACES'] + \
                "/{0}/log.json".format(ws_name)
            if os.path.isfile(ws_json):
                raw_logs = utils.reading_json(ws_json)

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

    # return all commands in flat
    @jwt_required
    def post(self, workspace):
        ws_name = utils.get_workspace(workspace=workspace)
        options_path = current_path + \
            '/storages/{0}/options.json'.format(ws_name)
        self.options = utils.reading_json(options_path)
        
        module = request.args.get('module')

        ws_name = os.path.basename(os.path.normpath(workspace))
        ws_name_encode = utils.url_encode(ws_name)

        utils.print_debug(ws_name)


        if ws_name in os.listdir(self.options['WORKSPACES']):
            ws_json = self.options['WORKSPACES'] + "/{0}/log.json".format(ws_name)
            raw_logs = utils.reading_json(ws_json)

        elif ws_name_encode in os.listdir(self.options['WORKSPACES']):
            ws_json = self.options['WORKSPACES'] + "/{0}/log.json".format(utils.url_encode(ws_name))
            # utils.print_debug(ws_json_encode)
            raw_logs = utils.reading_json(ws_json)

        if raw_logs:
            all_commands = []

            for k in raw_logs.keys():
                for item in raw_logs[k]:
                    cmd_item = item
                    cmd_item["module"] = k
                    cmd_item['std_path'] = utils.replace_argument(
                        self.options, item.get('std_path')).replace(self.options['WORKSPACES'], '')
                    cmd_item['output_path'] = utils.replace_argument(
                        self.options, item.get('output_path')).replace(self.options['WORKSPACES'], '')
                    cmd_item["module"] = k
                    all_commands.append(cmd_item)

            return {"commands": all_commands}
        else:
            return {"error": "Not found logs file for {0} workspace".format(ws_name)}
