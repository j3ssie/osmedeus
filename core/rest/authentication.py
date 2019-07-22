
import os
import json
import glob
import datetime

from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import (
    JWTManager, jwt_required, create_access_token,
    get_jwt_identity
)

from .decorators import local_only
import utils

'''
Check authentication
'''

current_path = os.path.dirname(os.path.realpath(__file__))


class Authentication(Resource):
    parser = reqparse.RequestParser()
    parser.add_argument('username',
                        type=str,
                        required=True,
                        help="This field cannot be left blank!"
                        )
    parser.add_argument('password',
                        type=str,
                        required=True,
                        help="This field cannot be left blank!"
                        )

    # add another authen level when settings things from remote
    def verify(self, options):
        config_path = options.get('CONFIG_PATH')
        if config_path:
            # get cred from config file
            config = ConfigParser(interpolation=ExtendedInterpolation())
            config.read(config_path)
            config_username = config['Server']['username']
            config_password = config['Server']['password']

            if config_username.lower() == options.get('USERNAME').lower() and config_password.lower() == options.get('PASSWORD').lower():
                return True

        return False

    # just look for right cred on any workspace
    def get_options(self, username, password):
        option_files = glob.glob(
            current_path + '/storages/**/options.json', recursive=True)
        # loop though all options avalible
        for option in option_files:
            json_option = utils.reading_json(option)
            if username == json_option.get('USERNAME'):
                if password == json_option.get('PASSWORD'):
                    return True
        return False
    
    # @local_only
    def post(self, workspace=None):
        # global options
        data = Authentication.parser.parse_args()
        username = data['username']
        password = data['password']

        # if no workspace specific
        if not workspace:
            if self.get_options(username, password):
                # cause we don't have real db so it's really hard to manage JWT
                # just change the secret if you want to revoke old token
                expires = datetime.timedelta(days=365)
                token = create_access_token(username, expires_delta=expires)
                return {'access_token': token}
            else:
                return {'error': "Credentials Incorrect"}
        elif workspace == 'None':
            pass

        current_path = os.path.dirname(os.path.realpath(__file__))

        options_path = current_path + \
            '/storages/{0}/options.json'.format(workspace)

        if not utils.not_empty_file(options_path):
            return {'error': "Workspace not found"}

        options = utils.reading_json(options_path)

        if username == options.get('USERNAME'):
            if password == options.get('PASSWORD'):
                # cause we don't have real db so it's really hard to manage JWT
                # just change the secret if you want to revoke old token
                expires = datetime.timedelta(days=365)
                token = create_access_token(username, expires_delta=expires)
                return {'access_token': token}
        
        return {'error': "Credentials Incorrect"}
