from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers
from api.models import Commands, Activities, Workspaces, Logs

# import utils
from core import utils
from core import common
from core import execute


def parse_data(data):
    item = {
        'cmd': data.get('cmd'),
        'output_path': data.get('output_path'),
        'std_path': data.get('std_path'),
        'module': data.get('module'),
        'status': data.get('status'),
        'cmd_type': data.get('cmd_type'),
        'workspace': data.get('workspace'),
        'resources': data.get('resources'),
        'chunk': data.get('chunk'),
        'delay': data.get('delay'),
    }
    return item


class ExecuteSerializer(serializers.Serializer):
    cmd = serializers.CharField(default='')
    output_path = serializers.CharField(allow_blank=True, default='')
    std_path = serializers.CharField(allow_blank=True, default='')
    # alias = serializers.CharField(default=None)
    workspace = serializers.CharField(required=True)
    module = serializers.CharField(default='default')
    cmd_type = serializers.CharField(default='single')
    forced = serializers.BooleanField(default=False)
    nolog = serializers.BooleanField(default=False)
    resources = serializers.CharField(allow_blank=True, default=False)
    chunk = serializers.IntegerField(default=1)
    delay = serializers.IntegerField(default=1)


class ExecuteView(APIView):
    serializer_class = ExecuteSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        serializer = ExecuteSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.data
        forced = data.get('forced')
        cmd = data.get('cmd')
        output_path = data.get('output_path')
        cmd_type = data.get('cmd_type')
        # don't care about the status
        nolog = data.get('nolog')

        # forced check if output path exist
        if not forced:
            if utils.not_empty_file(output_path):
                return common.message(500, "Commands is already done")

        # set idle status
        if nolog:
            data['status'] = 'Done'
        else:
            data['status'] = 'Running'

        item = parse_data(data)
        instance = Activities.objects.create(**item)
        Logs.objects.create(**item)

        command_record = instance.as_json()

        # really run the command
        if cmd_type == 'single':
            utils.print_info("Execute: {0} ".format(cmd))
            execute.run_single(command_record)
        elif cmd_type == 'list':
            utils.print_info("Execute chunk: {0} ".format(cmd))
            commands = execute.get_chunk_commands(command_record)
            execute.run_chunk(commands, command_record.get(
                'chunk'), command_record.get('delay'))

        # update status after done
        if instance.status != 'Done':
            instance.status = 'Done'
            instance.save()

        return common.message(200, "Commands is done")
