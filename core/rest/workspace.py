import os, json, glob
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
import utils

'''
Workspace listing and detail
'''

current_path = os.path.dirname(os.path.realpath(__file__))


class Workspaces(Resource):
    @jwt_required
    def get(self):
        # get all options file availiable
        option_files = glob.glob(current_path + '/storages/**/options.json', recursive=True)
        if not option_files:
            return {'error': 'No worksapce avaliable'}

        ws = []
        for options in option_files:
            json_options = utils.reading_json(options)
            ws.extend([ws for ws in os.listdir(
                json_options['WORKSPACES']) if ws[0] != '.'])

        return {'workspaces': list(set(ws))}



# get main json by workspace name
class Workspace(Resource):

    @jwt_required
    def get(self, workspace):
        #
        # @TODO potential LFI here
        #
        ws_name = os.path.basename(os.path.normpath(workspace))
        options_path = current_path + \
            '/storages/{0}/options.json'.format(ws_name)
        self.options = utils.reading_json(options_path)

        if ws_name in os.listdir(self.options['WORKSPACES']):
            ws_json = self.options['WORKSPACES'] + "/{0}/{0}.json".format(ws_name)
            if os.path.isfile(ws_json):
                utils.reading_json(ws_json)
                return utils.reading_json(ws_json)
        return 'Custom 404 here', 404
