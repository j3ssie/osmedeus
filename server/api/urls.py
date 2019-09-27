"""
rest URL Configuration
"""
from django.contrib import admin
from django.conf.urls import url, include
from django.contrib.staticfiles.urls import staticfiles_urlpatterns

from api.v1.execute.views import ExecuteView
from api.v1.commands.views import CommandsView, DetailCommandsDetailView
from api.v1.configs.views import ConfigurationsView
from api.v1.workspaces.views import WorkspacesView, WorkspacesListView
from api.v1.options.views import OptionsView
from api.v1.activities.views import ActivitiesView
from api.v1.logs.views import LogsView
from api.v1.reports.views import ReportsSkeletonView, ReportsView
from api.v1.monitors.views import MonitorsView
from api.v1.exports.views import ExportSumView
from api.v1.stdout.views import StdOutView
from api.v1.summaries.views import (
    SummariesView, 
    SummariesListView, 
    SummariesFieldView
)
from api.v1.clear.views import (
    ClearSummaries,
    ClearActivities,
    ClearSpecificActivities,
    ClearConfigurations,
    ClearWorkspaces,
)


urlpatterns = [
    # url(r'^api/cmd/$', PingView.as_view(), name='ping'),

    # Load commands
    url(r'^api/cmd/load/$', CommandsView.as_view(),
        name='load_command'),

    # Load Config
    url(r'^api/config/load/$', ConfigurationsView.as_view(),
        name='load_stateless_config'),

    # Create workspace record
    url(r'^api/workspaces/$', WorkspacesListView.as_view(),
        name='create_workspace'),

    # Create workspace record
    url(r'^api/workspace/create/$', WorkspacesView.as_view(),
        name='create_workspace'),

    # Get commands
    url(r'^api/monitor/$', MonitorsView.as_view(),
        name='monitor'),

    # Get workspace config
    url(r'^api/workspace/get/$', OptionsView.as_view(),
        name='get_options'),

    # Get commands
    url(r'^api/commands/get/$', DetailCommandsDetailView.as_view(),
        name='get_commands'),


    # Get activities
    url(r'^api/activities/get/$', ActivitiesView.as_view(),
        name='get_commands'),
    url(r'^api/logs/get/$', LogsView.as_view(),
        name='get_commands'),
    url(r'^api/activities/clear/$', ClearSpecificActivities.as_view(),
        name='get_commands'),

    # Get report skeleton
    url(r'^api/reports/raw/$', ReportsSkeletonView.as_view(),
        name='get_commands'),

    # Get report csv
    url(r'^api/exports/csv/$', ExportSumView.as_view(),
        name='get_commands'),

    # Get std output
    url(r'^api/stdout/get/$', StdOutView.as_view(),
        name='get_stdout'),

    # Get real report
    url(r'^api/reports/real/$', ReportsView.as_view(),
        name='get_commands'),

    # Get Summary list
    url(r'^api/summaries/set/$', SummariesListView.as_view(),
        name='set_list_domains'),

    url(r'^api/summaries/get/$', SummariesView.as_view(),
        name='set_single_domains'),

    url(r'^api/summaries/field/$', SummariesFieldView.as_view(),
        name='set_single_domains'),

    # Execute command
    url(r'^api/cmd/execute/$', ExecuteView.as_view(),
        name='execute'),

    # Clear table
    url(r'^api/clear/summaries/$', ClearSummaries.as_view(),
        name='execute'),
    url(r'^api/clear/activities/$', ClearActivities.as_view(),
        name='execute'),
    url(r'^api/clear/configs/$', ClearConfigurations.as_view(),
        name='execute'),
    url(r'^api/clear/workspaces/$', ClearWorkspaces.as_view(),
        name='execute'),


]
urlpatterns += staticfiles_urlpatterns()
