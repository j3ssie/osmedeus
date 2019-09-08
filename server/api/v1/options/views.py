from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers

# from api.models import Options

from core import utils
from core import dbutils
from core import common

'''
Endpoint to get options
'''


# define field
class OptionsSerializer(serializers.Serializer):
    workspace = serializers.CharField(required=True)
    # override = serializers.BooleanField(default=False)


class OptionsView(APIView):
    serializer_class = OptionsSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        serializer = OptionsSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.data

        workspace = data.get('workspace')
        options = dbutils.get_stateful_options(workspace)

        if not options:
            return common.message(404, "Workspace not found")

        # get absolute path for easy parse command
        real_workspace = utils.join_path(options.get(
            'WORKSPACES'), options.get('WORKSPACE'))
        options['WORKSPACE'] = real_workspace

        return common.returnJSON(options)
