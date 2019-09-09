# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import models
from core import utils
# from django.db.models import signals
# from django.dispatch import receiver


# store some configuration
class Configurations(models.Model):
    name = models.TextField(unique=True, blank=False, default='')
    value = models.TextField(blank=False, default='')
    alias = models.TextField(default='')
    desc = models.TextField(default='')

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            name=self.name,
            value=self.value,
            alias=self.alias,
            desc=self.desc,
        )


# sekeleton of the command
class Commands(models.Model):
    cmd = models.TextField(blank=False)
    requirement = models.TextField(default='')
    output_path = models.TextField(default='')
    std_path = models.TextField(default='')
    mode = models.TextField(default='general')
    cmd_type = models.TextField(default='single')
    speed = models.TextField(default='quick')
    pre_run = models.TextField(default='')
    post_run = models.TextField(default='')
    waiting = models.TextField(default='')
    cleaned_output = models.TextField(default='')
    chunk = models.IntegerField(default=1)
    delay = models.IntegerField(default=1)
    banner = models.TextField(default='')
    resources = models.TextField(default='')
    alias = models.TextField(default='')
    module = models.TextField(default='general')
    note = models.TextField(default='')
    checksum = models.TextField(
        unique=True, blank=False,
        default='e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855')

    def save(self, *args, **kwargs):
        # prevent duplicate analyzed command
        raw_check = str(self.cmd) + \
            str(self.output_path) + str(self.std_path) + \
            str(self.alias) + str(self.cmd_type) + \
            str(self.mode) + str(self.pre_run) + \
            str(self.resources) + str(self.speed)
        self.checksum = utils.gen_checksum(raw_check)
        try:
            super(Commands, self).save(*args, **kwargs)
        except Exception:
            utils.print_bad("Duplicate Commands")

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            cmd=self.cmd,
            requirement=self.requirement,
            output_path=self.output_path,
            std_path=self.std_path,
            alias=self.alias,
            module=self.module,
            cmd_type=self.cmd_type,
            chunk=self.chunk,
            delay=self.delay,
            banner=self.banner,
            resources=self.resources,
            pre_run=self.pre_run,
            post_run=self.post_run,
            waiting=self.waiting,
            cleaned_output=self.cleaned_output,
        )


# place to check if command is done or not
class Activities(models.Model):
    cmd = models.TextField(blank=False)
    output_path = models.TextField(default='')
    std_path = models.TextField(default='')
    module = models.TextField(default='general')
    status = models.TextField(default=False)
    workspace = models.TextField(default=False)
    cmd_type = models.TextField(default='single')
    chunk = models.IntegerField(default=1)
    delay = models.IntegerField(default=1)
    resources = models.TextField(default='')

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            cmd=self.cmd,
            output_path=self.output_path,
            std_path=self.std_path,
            module=self.module,
            status=self.status,
            workspace=self.workspace,
            chunk=self.chunk,
            delay=self.delay,
            resources=self.resources,
        )


# place to check if command is done or not
class Logs(models.Model):
    cmd = models.TextField(blank=False)
    output_path = models.TextField(default='')
    std_path = models.TextField(default='')
    module = models.TextField(default='general')
    workspace = models.TextField(default=False)
    cmd_type = models.TextField(default='single')
    status = models.TextField(default=False)
    chunk = models.IntegerField(default=1)
    delay = models.IntegerField(default=1)
    resources = models.TextField(default='')
    checksum = models.TextField(
        unique=True, blank=False,
        default='e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855')

    def save(self, *args, **kwargs):
        raw_check = str(self.cmd) + \
            str(self.output_path) + str(self.std_path) + \
            str(self.workspace) + str(self.cmd_type)
        self.checksum = utils.gen_checksum(raw_check)
        try:
            super(Logs, self).save(*args, **kwargs)
        except Exception:
            pass

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            cmd=self.cmd,
            output_path=self.output_path,
            std_path=self.std_path,
            module=self.module,
            workspace=self.workspace,
            chunk=self.chunk,
            status=self.status,
            delay=self.delay,
            resources=self.resources,
            checksum=self.checksum,
        )


class Workspaces(models.Model):
    raw_target = models.TextField(blank=False)
    target = models.TextField(default='')
    ip_address = models.TextField(default='')
    workspace = models.TextField(blank=False, unique=True)
    output = models.TextField(blank=False)
    mode = models.TextField(default='')
    workspaces = models.TextField(default='')
    # list modules will run in the routine
    modules = models.TextField(default='')
    speed = models.TextField(default='quick|*;;slow|-')
    # where to run command and store result
    arch = models.TextField(default='server')
    forced = models.BooleanField(default=False)
    verbose = models.BooleanField(default=False)
    created_time = models.IntegerField(default=0)

    def save(self, *args, **kwargs):
        self.created_time = utils.gen_ts()
        try:
            super(Workspaces, self).save(*args, **kwargs)
        except Exception:
            utils.print_bad("Duplicate workspace or missing Field")

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            raw_target=self.raw_target,
            target=self.target,
            ip_address=self.ip_address,
            workspace=self.workspace,
            output=self.output,
            workspaces=self.workspaces,
            mode=self.mode,
            modules=self.modules,
            speed=self.speed,
            arch=self.arch,
            forced=self.forced,
            verbose=self.verbose,
        )


class Summaries(models.Model):
    domain = models.TextField(default='')
    ip_address = models.TextField(default='N/A')
    technologies = models.TextField(default='N/A')
    ports = models.TextField(default='N/A')
    workspace = models.TextField(blank=False)
    paths = models.TextField(default='')
    screenshot = models.TextField(default='')
    note = models.TextField(default='')
    checksum = models.TextField(
        unique=True, blank=False,
        default='e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855')

    def save(self, *args, **kwargs):
        if self.domain is None:
            self.domain = self.ip_address
        # prevent duplicate analyzed command
        raw_check = str(self.domain) + str(self.workspace) 
        self.checksum = utils.gen_checksum(raw_check)
        try:
            super(Summaries, self).save(*args, **kwargs)
        except Exception:
            utils.print_bad("Duplicate Summaries Record")

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            domain=self.domain,
            ip_address=self.ip_address,
            technologies=self.technologies,
            ports=self.ports,
            workspace=self.workspace,
            paths=self.paths,
            screenshot=self.screenshot,
            note=self.note,
            checksum=self.checksum,
        )


class ReportsSkeleton(models.Model):
    report_path = models.TextField(default='')
    report_type = models.TextField(default='')
    module = models.TextField(default='')
    note = models.TextField(default='')
    mode = models.TextField(default='general')

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            report_path=self.report_path,
            report_type=self.report_type,
            module=self.module,
            note=self.note,
            mode=self.mode,
        )


class Reports(models.Model):
    report_path = models.TextField(default='')
    report_type = models.TextField(default='')
    note = models.TextField(default='')
    module = models.TextField(default='')
    workspace = models.TextField(default='')
    mode = models.TextField(default='general')

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            report_path=self.report_path,
            report_type=self.report_type,
            note=self.note,
            module=self.module,
            mode=self.mode,
            workspace=self.workspace,
        )


class Exploits(models.Model):
    description = models.TextField(default='')
    condition_command = models.TextField(default='')
    condition_tech = models.TextField(default='')
    condition_content = models.TextField(default='')
    exploit_command = models.TextField(default='')

    def as_json(self):
        return dict(
            source_model=self.__class__.__name__,
            pk=self.pk,
            description=self.description,
            condition_command=self.condition_command,
            condition_tech=self.condition_tech,
            condition_content=self.condition_content,
            exploit_command=self.exploit_command,
        )
