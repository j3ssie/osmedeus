from rest_framework import generics, mixins
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers
from rest_framework import filters

from api.models import Summaries

from core import utils
from core import dbutils
from core import common


class SummariesListSerializer(serializers.Serializer):
    domains = serializers.ListField(
        child=serializers.CharField(), allow_empty=True)
    domains_file = serializers.CharField(allow_blank=True, default=False)
    workspace = serializers.CharField(required=True)
    update_type = serializers.CharField(default='partial')


class SummariesListView(generics.ListAPIView, APIView):
    serializer_class = SummariesListSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        serializer = SummariesListSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.data

        workspace = data.get('workspace')
        domains = data.get('domains')
        domains_file = data.get('domains_file')
        update_type = data.get('update_type')

        if not domains:
            # @TODO Check if file under workspaces folder to prevent a LFI
            domains = utils.just_read(domains_file, get_list=True)
        if not domains:
            return common.message(500, "Domain files not found")

        # create domain record
        for domain in domains:
            jsonl = dbutils.parse_domains(domain)
            dbutils.import_domain_summary(jsonl,  workspace, update_type)
        return common.message(200, "Summary List Submitted")


class SummariesSerializer(serializers.ModelSerializer):
    class Meta:
        model = Summaries
        fields = [
            'domain',
            'ip_address',
            'technologies',
            'ports',
            'workspace',
            'paths',
            'screenshot',
            'note',
            'checksum',
        ]


class SummariesView(
        generics.ListAPIView,
        APIView):

    search_fields = ['domain',
                     'ip_address',
                     'technologies',
                     'ports',
                     'workspace',
                     'paths',
                     'screenshot',
                     'note',
                     'checksum']
    serializer_class = SummariesSerializer
    filter_backends = (filters.SearchFilter,)
    permission_classes = (IsAuthenticated, IsAdminUser)

    def get_queryset(self):
        queryset = Summaries.objects.all()
        workspace = self.request.query_params.get('workspace', None)
        if workspace is not None:
            queryset = queryset.filter(workspace=workspace)
        return queryset

    def get(self, request, *args, **kwargs):
        return Response({'summaries': self.list(request, *args, **kwargs).data})


class SummariesFieldView(APIView):
    permission_classes = (IsAuthenticated, IsAdminUser)

    def get(self, request, *args, **kwargs):
        workspace = request.query_params.get('workspace', None)
        if not workspace:
            return common.message(500, "Workspace not found")

        field = request.query_params.get('field', None)
        data = []
        if 'ip' in field:
            data = list(Summaries.objects.filter(workspace=workspace).values_list('ip_address', flat=True).exclude(ip_address='N/A'))

        return Response({'summaries': data})
