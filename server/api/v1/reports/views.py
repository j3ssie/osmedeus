from rest_framework import generics, mixins
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework.permissions import IsAuthenticated, IsAdminUser
from rest_framework import serializers

from api.models import ReportsSkeleton, Reports, Workspaces

from core import utils
from core import dbutils
from core import common


class ReportsSkeletonSerializer(serializers.ModelSerializer):
    class Meta:
        model = ReportsSkeleton
        fields = [
            'report_path',
            'report_type',
            'module',
            'note',
            'mode',
        ]


class ReportsSkeletonView(
        generics.ListAPIView,
        APIView):

    serializer_class = ReportsSkeletonSerializer
    permission_classes = (IsAuthenticated, IsAdminUser)

    def get_queryset(self):
        queryset = ReportsSkeleton.objects.all()
        module = self.request.query_params.get('module', None)
        if module is not None:
            queryset = queryset.filter(module=module)

        note = self.request.query_params.get('note', None)
        if note is not None:
            queryset = queryset.filter(note__contains=note)

        return queryset

    def get(self, request, *args, **kwargs):
        return Response({'reports': self.list(request, *args, **kwargs).data})


# Real report
class ReportsView(APIView):
    permission_classes = (IsAuthenticated, IsAdminUser)

    def get_reports(self, options, module=None, full=False, grouped=True):
        queryset = ReportsSkeleton.objects.all()
        if module is not None:
            queryset = queryset.filter(module=module)
            modules = [module]
        else:
            modules = list(ReportsSkeleton.objects.values_list(
                'module', flat=True).distinct())

        group_report = [{'module': m, 'reports': []} for m in modules]

        reports = []
        for record in queryset:
            report = record.as_json()
            report_path = utils.replace_argument(options, report.get('report_path'))
            # print(report_path)
            if utils.not_empty_file(report_path):
                if full:
                    report['report_path'] = report_path.replace(options.get('WORKSPACES'), '')
                else:
                    report['report_path'] = report_path.replace(options.get('WORKSPACES'), '').strip('/')
                reports.append(report)

        if not grouped:
            return reports

        seen = []
        for i in range(len(group_report)):
            for report in reports:
                if report.get('module') == group_report[i]['module']:
                    if report.get('report_path') not in seen:
                        group_report[i]['reports'].append(report)
                        seen.append(report.get('report_path'))

        return group_report

    def get(self, request, *args, **kwargs):
        workspace = request.query_params.get('workspace', None)
        module = request.query_params.get('module', None)
        full = request.query_params.get('full', None)
        grouped = request.query_params.get('grouped', None)
        # workspace validate
        if not workspace or workspace == 'null':
            return common.message(500, "Workspace not specifed")
        obj = Workspaces.objects.filter(workspace=workspace)
        if not obj.first():
            return common.message(404, "Workspace not found")

        # get options
        ws = obj.first().as_json().get('workspace')
        options = dbutils.get_stateful_options(ws)
        real_workspace = utils.join_path(options.get(
            'WORKSPACES'), options.get('WORKSPACE'))
        options['WORKSPACE'] = real_workspace
        reports = self.get_reports(options, module, full, grouped)

        content = {'reports': reports}
        return Response(content)

