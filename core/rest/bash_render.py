import os
import glob
import json
from pathlib import Path
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
from flask import Flask, request, escape, make_response
import utils

# incase you can't install ansi2html it's won't break the api
try:
    from ansi2html import Ansi2HTMLConverter
except:
    pass

current_path = os.path.dirname(os.path.realpath(__file__))

'''
render stdout content 
'''

class BashRender(Resource):

    def verify_file(self, filename):
        option_files = glob.glob(
            current_path + '/storages/**/options.json', recursive=True)
        # loop though all options avalible
        for option in option_files:
            json_option = utils.reading_json(option)
            stdout_path = json_option.get('WORKSPACES') + "/" + filename

            if utils.not_empty_file(stdout_path):
                return stdout_path
            # get real path
            p = Path(filename)
            ws = p.parts[0]
            if ws != utils.url_encode(ws):
                # just replace the first one
                filename_encode = filename.replace(ws, utils.url_encode(ws), 1)
                stdout_path = json_option.get('WORKSPACES') + "/" + filename_encode

                if utils.not_empty_file(stdout_path):
                    return stdout_path

        return False

    def get(self, filename):

        stdout_path = self.verify_file(filename)
        if not stdout_path:
            return 'Custom 404 here', 404
        
        content = utils.just_read(stdout_path).replace("\n\n", "\n")
        # content = utils.just_read(stdout_path)

        try:
            #convert console output to html
            conv = Ansi2HTMLConverter(
                scheme='mint-terminal')
            # conv = Ansi2HTMLConverter()
            html = conv.convert(content)
            response = make_response(html)
            response.headers['content-type'] = 'text/html'
            return response
        except:
            return content
