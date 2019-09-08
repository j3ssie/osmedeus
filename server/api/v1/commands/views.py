from rest_framework import generics, mixins
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers

from api.models import Commands

from core import utils
from core import dbutils
from core import common


# define field
class CommandsSerializer(serializers.Serializer):
    command_path = serializers.CharField(required=False, allow_blank=True)
    override = serializers.BooleanField(default=False)
    reset = serializers.BooleanField(default=False)


class CommandsView(APIView):
    serializer_class = CommandsSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        serializer = CommandsSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.data

        if data.get('override'):
            utils.print_good(
                "Clean all record in Commands")
            Commands.objects.all().delete()
        if data.get('reset'):
            dbutils.internal_parse_commands()
            return common.message(200, "Load default Commands Successfully")

        command_path = data.get('command_path')
        result = dbutils.parse_commands(command_path)
        if result:
            return common.message(200, "Parsing Commands Successfully")

        return common.message(500, "Commands file not exist")


class DetailCommandsSerializer(serializers.ModelSerializer):
    class Meta:
        model = Commands
        fields = [
            'cmd',
            'output_path',
            'std_path',
            'mode',
            'cmd_type',
            'speed',
            'requirement',
            'pre_run',
            'post_run',
            'waiting',
            'cleaned_output',
            'chunk',
            'delay',
            'banner',
            'resources',
            'alias',
            'module',
            'checksum',
        ]


class DetailCommandsDetailView(
        generics.ListAPIView,
        APIView):

    serializer_class = DetailCommandsSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)
    # queryset = Traffic.objects.all()
    # lookup_field = 'id'

    def get_queryset(self):
        queryset = Commands.objects.all()
        module = self.request.query_params.get('module', None)
        mode = self.request.query_params.get('mode', None)
        if module is not None:
            queryset = queryset.filter(module=module)

        if mode is not None:
            queryset = queryset.filter(mode=mode)

        alias = self.request.query_params.get('alias', None)
        if alias is not None:
            queryset = queryset.filter(alias__contains=alias)

        return queryset

    def get(self, request, *args, **kwargs):
        return Response({'commands': self.list(request, *args, **kwargs).data})
