import os, json
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
import utils
current_path = os.path.dirname(os.path.realpath(__file__))

'''
workspace listing and detail
'''

class Workspaces(Resource):
    # just return list of workspaces
    def __init__(self, **kwargs):
        self.options = utils.reading_json(current_path + '/storages/options.json')

    @jwt_required
    def get(self):
        # just remove hidden file
        ws = [ws for ws in os.listdir(self.options['WORKSPACES']) if ws[0] != '.']
        return {'workspaces': ws}



#get main json by workspace name
class Workspace(Resource):
    def __init__(self, **kwargs):
        self.options = utils.reading_json(current_path + '/storages/options.json')

    @jwt_required
    def get(self, workspace):
        #
        # @TODO potential LFI here
        #
        ws_name = os.path.basename(os.path.normpath(workspace))
        
        if ws_name in os.listdir(self.options['WORKSPACES']):
            ws_json = self.options['WORKSPACES'] + "/{0}/{0}.json".format(ws_name)
            if os.path.isfile(ws_json):
                utils.reading_json(ws_json)
                return utils.reading_json(ws_json)
        return 'Custom 404 here', 404
