from rest_framework import generics, mixins
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers

from api.models import Logs, Workspaces

from core import utils
from core import dbutils
from core import common


class LogsSerializer(serializers.ModelSerializer):
    class Meta:
        model = Logs
        fields = [
            'cmd',
            'output_path',
            'std_path',
            'module',
            'workspace',
            'cmd_type',
            'chunk',
            'delay',
            'resources',
            'checksum',
        ]


class LogsView(
        generics.ListAPIView,
        APIView):

    serializer_class = LogsSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def get_queryset(self):
        queryset = Logs.objects.all()
        module = self.request.query_params.get('module', None)
        workspace = self.request.query_params.get('workspace', None)
        cmd = self.request.query_params.get('cmd', None)
        raw = self.request.query_params.get('raw', None)

        if workspace is not None:
            queryset = queryset.filter(workspace=workspace)

        if module is not None:
            queryset = queryset.filter(module=module)

        if cmd is not None:
            queryset = queryset.filter(cmd__contains=cmd)
        if raw:
            return queryset

        real_queryset = []
        for item in queryset:
            if utils.not_empty_file(item.output_path):
                real_queryset.append(item)
        return real_queryset

    def get(self, request, *args, **kwargs):
        workspace = self.request.query_params.get('workspace', None)
        obj = Workspaces.objects.get(workspace=workspace)
        wss = obj.workspaces

        content = self.list(request, *args, **kwargs).data
        response = []
        for log in content:
            item = dict(log)
            item['output_path'] = item['output_path'].replace(wss, '').strip('/')
            item['std_path'] = item['std_path'].replace(wss, '').strip('/')
            response.append(item)

        return Response({'logs': response})

