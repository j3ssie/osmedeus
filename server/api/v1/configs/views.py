from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers

from api.models import Configurations

from core import utils
from core import dbutils
from core import common


# define field
class ConfigurationsSerializer(serializers.Serializer):
    config_path = serializers.CharField(default=False)
    override = serializers.BooleanField(default=False)


class ConfigurationsView(APIView):
    serializer_class = ConfigurationsSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def post(self, request, *args, **kwargs):
        serializer = ConfigurationsSerializer(data=request.data)
        serializer.is_valid(raise_exception=True)
        data = serializer.data

        if data.get('override'):
            utils.print_good(
                "Clean all record in Configurations")
            Configurations.objects.all().delete()

        config_path = data.get('config_path')
        if not config_path or config_path == 'False':
            config_path = utils.DEAFULT_CONFIG_PATH

        result = dbutils.load_default_config(config_path)
        if result:
            return common.message(200, "Loading Configurations Successfully")

        return common.message(500, "Configurations file not exist")
