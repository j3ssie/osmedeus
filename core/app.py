import os
import json
import sys
import subprocess
import time
import logging
import argparse
from pprint import pprint


import execute
import slack
import utils

from flask import abort
from flask_cors import CORS

from flask import Flask, jsonify, render_template, request, send_from_directory
from flask_jwt_extended import (
    JWTManager, jwt_required, create_access_token,
    get_jwt_identity
)
from flask_restful import Api, Resource, reqparse

from rest.decorators import local_only

from rest.cmd import Cmd
from rest.authentication import Authentication
from rest.configuration import Configurations
from rest.workspace import Workspace, Workspaces
from rest.activities import Activities
from rest.logs import Logs
from rest.modules import Modules
from rest.routines import Routines
from rest.bash_render import BashRender
from rest.save import Save

current_path = os.path.dirname(os.path.realpath(__file__))
############
## Flask config stuff

app = Flask('Osmedeus')
app = Flask(__name__, template_folder='ui/', static_folder='ui/static/')
#just for testing whitelist your domain if you wanna run this server remotely
cors = CORS(app, resources={r"/*": {"origins": "*"}})
api = Api(app)

# setup jwt secret, make sure you change this!
app.config['JWT_SECRET_KEY'] = '-----BEGIN RSA PRIVATE KEY-----' # go ahead, spider
jwt = JWTManager(app)

############


# just turn off the server
def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()

@local_only
@app.route('/api/shutdown', methods=['POST'])
def shutdown():
    shutdown_server()
    return 'Server shutting down...'


api.add_resource(Configurations, '/api/config')
api.add_resource(Authentication, '/api/auth')
api.add_resource(Cmd, '/api/cmd')
api.add_resource(Activities, '/api/activities')
api.add_resource(Workspaces, '/api/workspace')
api.add_resource(Workspace, '/api/workspace/<string:workspace>')
api.add_resource(Logs, '/api/logs/<string:workspace>')
api.add_resource(Modules, '/api/module/<string:workspace>')
api.add_resource(Routines, '/api/routines')
api.add_resource(Save, '/api/save')
api.add_resource(BashRender, '/stdout/<path:filename>')


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
    if path != "" and os.path.exists(current_path + "/ui/" + path):
        return send_from_directory(current_path + '/ui/', path)
    else:
        return send_from_directory(current_path + '/ui/', 'index.html')
####

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument("-b", "--bind", action="store", dest='bind', default="127.0.0.1")
    parser.add_argument("-p", "--port", action="store", dest='port', default="5000")
    parser.add_argument("--debug", action="store_true", help='just for debug purpose')
    parser.add_argument("--nossl", action="store_true", help='Use plaintext')
    # turn on this you really want to run remote but I'm not recommend
    parser.add_argument("--remote", action="store_true", help='Allow bypass local protection decorators')
    args = parser.parse_args()

    host = str(args.bind)
    port = int(args.port)
    debug = args.debug

    if args.remote:
        print(" * Warning: You're allow to bypass local protection")
        app.config['REMOTE'] = True
    else:
        app.config['REMOTE'] = False

    if not args.debug:
        print(" * Logging: off")
        log = logging.getLogger('werkzeug')
        log.setLevel(logging.ERROR)

    #choose to use SSL or not
    if args.nossl:
        app.run(host=host, port=port, debug=debug)
    else:
        cert_path = current_path + '/certs/cert.pem'
        key_path = current_path + '/certs/key.pem'
        app.run(host=host, port=port, debug=debug, ssl_context=(cert_path, key_path))
