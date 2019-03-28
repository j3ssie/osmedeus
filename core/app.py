import os
import json
import sys
import subprocess
import time
import logging
from pprint import pprint

import execute
import slack
import utils

from flask import abort
from flask_cors import CORS

from flask import Flask, jsonify, render_template, request, send_from_directory
from flask_jwt import JWT, current_identity, jwt_required
from flask_restful import Api, Resource, reqparse

from rest.decorators import local_only

from rest.cmd import Cmd
from rest.configuration import Configurations
from rest.workspace import Workspace, Workspaces
from rest.activities import Activities
from rest.logs import Logs
from rest.modules import Modules
from rest.routines import Routines

current_path = os.path.dirname(os.path.realpath(__file__))
############
## Flask config 
##
## turn off the http log
log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)
##

app = Flask('Osmedeus')
#just for testing
app = Flask(__name__, static_folder='ui/static/')

cors = CORS(app, resources={r"/*": {"origins": "*"}})
api = Api(app)


############


# just turn off the server
def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()

@local_only
@app.route('/shutdown', methods=['POST'])
def shutdown():
    shutdown_server()
    return 'Server shutting down...'


api.add_resource(Configurations, '/api/config')
api.add_resource(Cmd, '/api/cmd')
api.add_resource(Activities, '/api/activities')
api.add_resource(Workspaces, '/api/workspace')
api.add_resource(Workspace, '/api/workspace/<string:workspace>')
api.add_resource(Logs, '/api/logs/<string:workspace>')
api.add_resource(Modules, '/api/module/<string:workspace>')
api.add_resource(Routines, '/api/routines')



#### serve HTML and image content
@app.route('/wscdn/<path:filename>')
def custom_static(filename):
    options = utils.reading_json(current_path + '/rest/storages/options.json')
    return send_from_directory(options['WORKSPACES'], filename)
#####


##### serve react build
@app.route('/', defaults={'path': ''})
@app.route('/<path:path>')
def serve(path):
    if path != "" and os.path.exists("ui/" + path):
        return send_from_directory('ui', path)
    else:
        return send_from_directory('ui', 'index.html')
####

if __name__ == '__main__':
    app.run(debug=True)  # important to mention debug=True

    # app.run(debug=False)  # important to mention debug=True
