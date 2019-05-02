
import os
import json
import datetime

from flask_restful import Api, Resource, reqparse
from flask_jwt_extended import (
    JWTManager, jwt_required, create_access_token,
    get_jwt_identity
)

from .decorators import local_only
import utils
'''
#set some config
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

    #@local_only
    def post(self):
        # global options
        data = Authentication.parser.parse_args()
        username = data['username']
        password = data['password']

        current_path = os.path.dirname(os.path.realpath(__file__))
        options = utils.reading_json(current_path + '/storages/options.json')

        if username == options.get('USERNAME'):
            if password == options.get('PASSWORD'):
                expires = datetime.timedelta(days=365)
                token = create_access_token(username, expires_delta=expires)
                return {'access_token': token}
        
        return {'error': "Credentials Incorrect"}
