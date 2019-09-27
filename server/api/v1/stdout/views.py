from rest_framework import generics, mixins
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers
from rest_framework import filters
from django.http import HttpResponse
from api.models import Summaries

import os
import pathlib
# incase you can't install ansi2html it's won't break the api
try:
    from ansi2html import Ansi2HTMLConverter
except:
    pass

from core import utils
from core import dbutils
from core import common


class StdOutView(APIView):
    # permission_classes = (IsAuthenticated, IsAdminUser)

    def get(self, request, *args, **kwargs):
        filename = self.request.query_params.get('std', None)
        is_html = self.request.query_params.get('html', False)
        if not filename:
            return common.message(500, 'stdout not found')

        options = dbutils.get_stateless_options()
        wss = options.get('WORKSPACES')

        filename = utils.clean_path(filename).strip('/')
        p = pathlib.Path(filename)

        # we don't want a LFI here even when we're admin
        if p.parts[0] not in os.listdir(wss):
            return common.message(500, 'Workspace not found ')

        stdout = utils.join_path(wss, filename)
        content = utils.just_read(stdout)
        if not content:
            return HttpResponse("No result found")

        if filename.endswith('.html') or is_html:
            return HttpResponse(content)

        try:
            # convert console output to html
            conv = Ansi2HTMLConverter(scheme='mint-terminal')
            html = conv.convert(content)
            return HttpResponse(html)
            # return Response(html)
        except:
            return HttpResponse(content)
