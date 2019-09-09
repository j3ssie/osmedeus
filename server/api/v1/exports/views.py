from rest_framework import generics, mixins
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers
from rest_framework import filters
from djqscsv import render_to_csv_response, write_csv
from api.models import Summaries

from core import utils
from core import dbutils
from core import common


class ExportsSerializer(serializers.Serializer):
    workspace = serializers.CharField(required=True)
    filename = serializers.CharField(allow_blank=True, default=False)


class ExportSumView(APIView):

    serializer_class = ExportsSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def get_queryset(self, workspace):
        queryset = Summaries.objects.all()
        # workspace = self.request.query_params.get('workspace', None)
        if workspace is not None:
            queryset = queryset.filter(workspace=workspace)
        return queryset

    def post(self, request, *args, **kwargs):
        serializer = ExportsSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.data
        workspace = data.get('workspace')
        filename = data.get('filename')
        qs = self.get_queryset(workspace)

        if filename == 'False' or not filename:
            return render_to_csv_response(qs)
        else:
            filename = utils.clean_path(filename)
            with open(filename, 'wb') as csv_file:
                write_csv(qs, csv_file)
            return common.message(200, filename)
