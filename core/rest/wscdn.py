import os
import glob
import json
from flask_restful import Resource, reqparse
from flask_jwt_extended import jwt_required
from flask import send_from_directory
import utils
from pathlib import Path
BASE_DIR = Path(os.path.dirname(os.path.abspath(__file__)))

# incase you can't install ansi2html it's won't break the api
try:
    from ansi2html import Ansi2HTMLConverter
except:
    pass

'''
render stdout content 
'''


class Wscdn(Resource):

    def verify_file(self, filename):
        option_files = glob.glob(
            str(BASE_DIR) + '/storages/**/options.json', recursive=True)

        # loop though all options avalible
        for option in option_files:
            json_option = utils.reading_json(option)
            stdout_path = json_option.get('WORKSPACES') + "/" + filename

            if utils.not_empty_file(stdout_path):
                return json_option.get('WORKSPACES'), os.path.normpath(filename)

            # get real path 
            p = Path(filename)
            ws = p.parts[0]
            if ws != utils.url_encode(ws):
                # just replace the first one
                filename_encode = filename.replace(ws, utils.url_encode(ws), 1)
                stdout_path_encode = json_option.get('WORKSPACES') + filename_encode
                if utils.not_empty_file(stdout_path_encode):
                    return json_option.get('WORKSPACES'), os.path.normpath(filename_encode)

        return False, False

    def get(self, filename):
        ws_path, stdout_path = self.verify_file(filename)

        if not stdout_path:
            return 'Custom 404 here', 404
        return send_from_directory(ws_path, stdout_path)
