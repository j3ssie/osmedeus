# -*- coding: utf-8 -*-

from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework import serializers
from rest_framework.permissions import IsAuthenticated, IsAdminUser

from api.models import *
from core import common


class ClearSummaries(APIView):
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        Summaries.objects.all().delete()
        return common.message(200, "Summaries Table have been cleared.")


class ClearSpecificSerializers(serializers.Serializer):
    workspace = serializers.CharField(required=True)
    module = serializers.CharField(required=True)


class ClearSpecificActivities(APIView):
    serializer_class = ClearSpecificSerializers
    permission_classes = (IsAuthenticated, IsAdminUser)

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

    # clear old activities
    def post(self, request, *args, **kwargs):
        serializer = ClearSpecificSerializers(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.data
        queryset = Activities.objects.all()
        queryset = queryset.filter(workspace=data.get('workspace'))
        queryset = queryset.filter(module=data.get('module'))

        queryset.delete()
        return common.message(200, "Clear all old activities for {0}")


class ClearActivities(APIView):
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        Activities.objects.all().delete()
        return common.message(200, "Activities Table have been cleared.")


class ClearConfigurations(APIView):
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        Configurations.objects.all().delete()
        return common.message(200, "Configurations Table have been cleared.")


class ClearWorkspaces(APIView):
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        Workspaces.objects.all().delete()
        return common.message(200, "Workspaces Table have been cleared.")

