import functools
import os
import sys
import subprocess
import time
import execute
import utils
import logging

from flask import Flask, jsonify, render_template, request
from flask_jwt import JWT, current_identity, jwt_required
from flask_restful import Api, Resource, reqparse


# # turn off the http log
# log = logging.getLogger('werkzeug')
# log.setLevel(logging.ERROR)

app = Flask(__name__)
# app = Flask(__name__, template_folder="templates/sample1/build/", static_folder="templates/sample1/build/static")
api = Api(app)

activities_log = {
}
processes = []

# only allow local executed


def local_only(f):
    @functools.wraps(f)
    def function_name(*args, **kwargs):
        src_ip = request.remote_addr
        if src_ip != "127.0.0.1":
            return "External Detected :("
        else:
            return f(*args, **kwargs)
    return function_name


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


class Cmd(Resource):
    parser = reqparse.RequestParser()
    parser.add_argument('cmd',
                        type=str,
                        required=True,
                        help="This field cannot be left blank!"
                        )

    parser.add_argument('output_path',
                        type=str,
                        required=True,
                        help="This field cannot be left blank!"
                        )

    parser.add_argument('std_path',
                        type=str,
                        required=False
                        )

    parser.add_argument('module',
                        type=str,
                        required=True
                        )

    @local_only
    def post(self):
        data = Cmd.parser.parse_args()
        cmd = data['cmd']
        std_path = data['std_path']
        output_path = data['output_path']
        module = data['module']

        activity = {
            'cmd': cmd,
            'std_path': std_path,
            'output_path': output_path,
            'status': 'Running'
        }

        # this
        if activities_log.get(module):
            activities_log[module].append(activity)
        else:
            activities_log[module] = [activity]

        utils.print_info("Execute: {0} ".format(cmd))
        stdout = execute.run(cmd)
        # just ignore for testing purpose
        # stdout = "<< stdoutput >> << {0} >>".format(cmd)
        if std_path != '':
            utils.just_write(std_path, stdout)
        utils.check_output(output_path)

        # change status of log
        # activity['status'] = 'Done'
        for item in activities_log[module]:
            if item['cmd'] == cmd:
                if stdout is None:
                    item['status'] = 'Error'
                item['status'] = 'Done'

        return jsonify(status="200", output_path=output_path)

# logging command


class Activity(Resource):
    parser = reqparse.RequestParser()
    parser.add_argument('cmd',
                        type=str,
                        required=False,
                        help="This field cannot be left blank!",
                        default=None
                        )

    # get all activity log or by module
    def get(self):
        # get specific module
        module = request.args.get('module')
        if module:
            cmds = activities_log[module]
            return {'commands': cmds}
        else:
            return activities_log

    def post(self):
        data = Activity.parser.parse_args()
        cmd = data['cmd']

        module = request.args.get('module')
        #if module avalible just ignore cmd stuff
        if module:
            if cmd:
                commands = [x for x in activities_log[module] if cmd in x['cmd']]
            else:
                commands = [x for x in activities_log[module]]
            return {'commands': commands}

        else:
            cmds = []
            for item in [x for x in list(activities_log.values())]:
                cmds += item
            commands = [x for x in cmds if cmd in x['cmd']]

            return {'commands': commands}

# reading report stuff
class Report(Resource):
    parser = reqparse.RequestParser()
    parser.add_argument('cmd',
                        type=str,
                        required=True,
                        help="This field cannot be left blank!"
                        )

    def get(self):
        pass


api.add_resource(Cmd, '/cmd')
api.add_resource(Activity, '/activities')
# api.add_resource(Report, '/report/<string:workspace>')
# api.add_resource(Report, '/report/<string:workspace>/<string:host>')


if __name__ == '__main__':
    app.run(debug=True)  # important to mention debug=True
