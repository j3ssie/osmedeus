from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers

from api.models import Workspaces

from core import utils
from core import dbutils
from core import common
# from routine import initials

'''
Endpoint to store parsed data in a record
'''


# define fields
class WorkspacesSerializer(serializers.Serializer):
    raw_target = serializers.CharField(required=True)
    target = serializers.CharField(required=False)
    workspace = serializers.CharField(required=False)
    output = serializers.CharField(required=False)

    workspaces = serializers.CharField(required=False)
    data_path = serializers.CharField(required=False)
    plugin_path = serializers.CharField(required=False)

    mode = serializers.CharField(default='normal')
    # list modules will run in the routine seperated by ','
    modules = serializers.CharField(required=False)
    arch = serializers.CharField(default='server')

    speed = serializers.CharField(default='quick|*;;slow|-')
    ip_address = serializers.CharField(default='')
    verbose = serializers.BooleanField(default=False)
    forced = serializers.BooleanField(default=False)


class WorkspacesView(APIView):
    serializer_class = WorkspacesSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        serializer = WorkspacesSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.data

        # just some verbose options
        mode = data.get('mode')
        verbose = data.get('verbose')
        speed = data.get('speed')
        forced = data.get('forced')
        arch = data.get('arch')

        # input part
        raw_target = data.get('raw_target')
        target = dbutils.clean_input(raw_target)
        # resolve IP if possible
        ip_address = utils.resolve_input(target)

        # strip slash for saving it as path
        workspace = utils.get_ws(target)

        # output and plugin part
        options = dbutils.get_stateless_options()
        workspaces = options.get('WORKSPACES')
        data_path = options.get('DATA_PATH')
        plugin_path = options.get('PLUGINS_PATH')

        modules = utils.set_value(
            dbutils.get_modules(mode), data.get('modules'))

        # set if defined or get it as default
        workspace = utils.set_value(workspace, data.get('workspace'))

        output = utils.set_value(workspace, data.get('output'))
        # data part
        workspaces = utils.set_value(workspaces, data.get('workspaces'))
        data_path = utils.set_value(data_path, data.get('data_path'))
        plugin_path = utils.set_value(plugin_path, data.get('plugin_path'))
        target = utils.set_value(target, data.get('target'))
        ip_address = utils.set_value(ip_address, data.get('ip_address'))

        # store it to db
        item = {
            'raw_target': raw_target,
            'target': target,
            'ip_address': ip_address,
            'workspace': workspace,
            'output': output,
            'workspaces': workspaces,
            'modules': modules,
            'arch': arch,
            'mode': mode,
            'speed': speed,
            'forced': forced,
            'verbose': verbose,
        }

        instance, created = Workspaces.objects.get_or_create(
            workspace=workspace)
        Workspaces.objects.filter(workspace=workspace).update(**item)
        real_workspace = utils.join_path(workspaces, workspace)

        if created:
            if instance.arch == 'server':
                utils.make_directory(real_workspace)
                return common.returnJSON({
                    "workspace": workspace,
                    "msg": "Workspaces created Successfully"
                }, 200)
        else:
            # person object already exists
            return common.returnJSON({
                "workspace": workspace,
                "msg": "Workspaces already exists"
            }, 442)


class WorkspacesListView(APIView):
    # serializer_class = WorkspacesSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def get(self, request, *args, **kwargs):
        workspaces = list(Workspaces.objects.all(
        ).values_list('workspace', flat=True))

        options = dbutils.get_stateless_options()
        wss = options.get('WORKSPACES')

        # remove blank workspace
        for ws in workspaces:
            real_ws = utils.join_path(wss, ws)
            if not utils.not_empty_dir(real_ws):
                workspaces.remove(ws)

        return common.returnJSON({
            "workspaces": workspaces,
        }, 200)
