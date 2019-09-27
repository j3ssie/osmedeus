from rest_framework import generics, mixins
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers

from api.models import Monitors, Workspaces

from core import utils
from core import dbutils
from core import common


class MonitorsViewSerializer(serializers.ModelSerializer):
    class Meta:
        model = Monitors
        fields = [
            'old_path',
            'new_path',
            'compare_ts',
            'diff_content',
            'diff_type',
            'workspace',
            'level',
            'notified',
        ]


class MonitorsView(mixins.CreateModelMixin, generics.ListAPIView, APIView):

    serializer_class = MonitorsViewSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def get_queryset(self):
        queryset = Monitors.objects.all()
        workspace = self.request.query_params.get('workspace', None)
        if workspace is not None:
            queryset = queryset.filter(workspace=workspace)
        return queryset

    def post(self, request, *args, **kwargs):
        return self.create(request, *args, **kwargs)

    def get(self, request, *args, **kwargs):
        content = self.list(request, *args, **kwargs).data
        return Response({'monitors': content})

