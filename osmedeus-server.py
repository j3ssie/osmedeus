#!/usr/bin/env python3

import os
import json
import logging
import argparse
from pprint import pprint
from pathlib import Path

from flask import Flask, request, send_from_directory
from flask_cors import CORS
from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import (
    JWTManager, jwt_required, create_access_token,
    get_jwt_identity
)

from Osmedeus.core import execute
from Osmedeus.core import utils
from Osmedeus.core.rest.decorators import local_only
from Osmedeus.core.rest.cmd import Cmd
from Osmedeus.core.rest.authentication import Authentication
from Osmedeus.core.rest.configuration import Configurations
from Osmedeus.core.rest.workspace import Workspace, Workspaces
from Osmedeus.core.rest.activities import Activities
from Osmedeus.core.rest.logs import Logs
from Osmedeus.core.rest.modules import Modules
from Osmedeus.core.rest.routines import Routines
from Osmedeus.core.rest.bash_render import BashRender
from Osmedeus.core.rest.wscdn import Wscdn
from Osmedeus.core.rest.save import Save
from Osmedeus.resources import *

OSMEDEUS_HOME = str(Path.home().joinpath('.osmedeus'))

# TODO: I think we need get it from the config file
STORAGES_DIR = OSMEDEUS_HOME + '/storages'
CERT_PATH = OSMEDEUS_HOME + '/certs/cert.pem'
KEY_PATH = OSMEDEUS_HOME + '/certs/key.pem'

# Flask config stuff
app = Flask(__name__, 
    template_folder=RESOURCES_PATH.joinpath('ui/'), 
    static_folder=RESOURCES_PATH.joinpath('ui/static/'))

# just for testing whitelist your domain if you wanna run this server remotely
cors = CORS(app, resources={r"/*": {"origins": "*"}})
api = Api(app)

# setup jwt secret
# SECURITY WARNING: change this if you running on remote
app.config['JWT_SECRET_KEY'] = '-----BEGIN RSA PRIVATE KEY-----'  # go ahead, spider
jwt = JWTManager(app)

'''
End of Flask config
'''

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


# API for the Osmedeus client
api.add_resource(Configurations, '/api/config')
api.add_resource(Authentication, '/api/auth', '/api/<string:workspace>/auth')
api.add_resource(Cmd, '/api/<string:workspace>/cmd')
api.add_resource(Activities, '/api/<string:workspace>/activities')
api.add_resource(Routines, '/api/<string:workspace>/routines')

# API for the web UI
api.add_resource(Modules, '/api/module/<string:workspace>')
api.add_resource(Workspaces, '/api/workspace')
api.add_resource(Logs, '/api/logs/<string:workspace>')
api.add_resource(Workspace, '/api/workspace/<string:workspace>')
api.add_resource(BashRender, '/stdout/<path:filename>')
api.add_resource(Save, '/api/save')
api.add_resource(Wscdn, '/wscdn/<path:filename>')


# serve react build
@app.route('/', defaults={'path': ''})
@app.route('/<path:path>')
def serve(path):
    if path != "" and os.path.exists(str(RESOURCES_PATH.joinpath('ui/' + path))):
        return send_from_directory(str(RESOURCES_PATH.joinpath('ui/')), path)
    else:
        return send_from_directory(str(RESOURCES_PATH.joinpath('ui/')), 'index.html')

# parsing some command from cli
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

    # make sure storages is created
    utils.make_directory(STORAGES_DIR)

    if args.remote:
        print(" * Warning: You're allow to bypass local protection")
        app.config['REMOTE'] = True
    else:
        app.config['REMOTE'] = False

    if not args.debug:
        print(" * Logging: off")
        log = logging.getLogger('werkzeug')
        log.setLevel(logging.ERROR)

    # choose to use SSL or not
    if args.nossl:
        app.run(host=host, port=port, debug=debug)
    else:
        if not os.path.exists(CERT_PATH) and not os.path.exists(KEY_PATH):
            print(" * WARNING: You're need to create your own cert in " + OSMEDEUS_HOME + "/certs")
            CERT_PATH = RESOURCES_PATH.joinpath('certs/cert.pem')
            KEY_PATH = RESOURCES_PATH.joinpath('certs/key.pem')
        app.run(host=host, port=port, debug=debug, ssl_context=(CERT_PATH, KEY_PATH))
