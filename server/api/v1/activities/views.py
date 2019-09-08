from rest_framework import generics, mixins
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers

from api.models import Activities

from core import utils
from core import dbutils
from core import common


class ActivitiesSerializer(serializers.ModelSerializer):
    class Meta:
        model = Activities
        fields = [
            'cmd',
            'output_path',
            'std_path',
            'module',
            'status',
            'workspace',
            'cmd_type',
            'chunk',
            'delay',
            'resources',
        ]


class ActivitiesView(generics.ListAPIView, APIView):
    serializer_class = ActivitiesSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)
    # lookup_field = 'id'

    def get_queryset(self):
        queryset = Activities.objects.all()
        workspace = self.request.query_params.get('workspace', None)
        if workspace is not None:
            queryset = queryset.filter(workspace=workspace)

        module = self.request.query_params.get('module', None)
        if module is not None:
            queryset = queryset.filter(module=module)

        cmd = self.request.query_params.get('cmd', None)
        if cmd is not None:
            queryset = queryset.filter(cmd__contains=cmd)

        return queryset

    def check_status(self, data):
        for item in data:
            if item.get('status') != 'Done':
                return 'Running'
        return 'Done'

    # update all to Done
    def post(self, request, *args, **kwargs):
        self.get_queryset().update(status='Done')
        return common.message(200, "Update all status")

    def get(self, request, *args, **kwargs):
        data = self.list(request, *args, **kwargs).data
        status = self.check_status(data)
        content = {
            'status': status,
            'activities': data,
        }
        return Response(content)
