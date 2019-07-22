import os
import json
import time
import hashlib
import urllib.parse
from flask_restful import Resource, reqparse
from flask_jwt_extended import jwt_required
from .decorators import local_only
import utils

'''
helper to store data into a file for quickly running from web UI
'''


class Save(Resource):
    parser = reqparse.RequestParser()
    parser.add_argument('content',
                        type=str,
                        required=True,
                        help="This field cannot be left blank!"
                        )

    @jwt_required
    def post(self):
        utils.make_directory('/tmp/osmedeus-tmp/')
        data = Save.parser.parse_args()
        raw_content = data['content']
        content = urllib.parse.unquote(raw_content)
        ts = str(int(time.time()))
        filepath = '/tmp/osmedeus-tmp/' + \
            hashlib.md5(ts.encode()).hexdigest()[:5]

        utils.just_write(filepath, content)
        return {"filepath": filepath}
