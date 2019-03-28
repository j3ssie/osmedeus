import functools
from flask import Flask, jsonify, render_template, request

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


