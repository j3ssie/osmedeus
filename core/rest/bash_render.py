import os
import json
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import jwt_required
from flask import Flask, request, escape, make_response
import utils
#incase you can't install ansi2html it's won't break the api
try:
    from ansi2html import Ansi2HTMLConverter
except:
    pass

current_path = os.path.dirname(os.path.realpath(__file__))

'''
render stdout content 
'''

class BashRender(Resource):

    def __init__(self, **kwargs):
        self.options = utils.reading_json(
            current_path + '/storages/options.json')

    def get(self, filename):
        # @TODO potential LFI here
        std_file = os.path.normpath(filename)
        stdout_path = self.options['WORKSPACES'] + std_file
        if not utils.not_empty_file(stdout_path):
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
