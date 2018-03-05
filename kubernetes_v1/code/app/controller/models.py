from django.db import models
from django.contrib.auth.models import User
from django.contrib.auth.admin import UserAdmin
from django.utils.translation import ugettext_lazy as _
from django.db.models.signals import post_save, pre_delete
from django.dispatch import receiver
from django.conf import settings
from django.core.validators import MaxValueValidator, MinValueValidator
import logging
import traceback
import re
import glob
import shutil
import os
from k8s_api import k8s

# Create your models here.


logger = logging.getLogger('servers')


class CustomUserAdmin(UserAdmin):
    fieldsets = (
        (None, {'fields': ('username', 'password')}),
        (_('Personal info'), {'fields': ('first_name', 'last_name', 'email')}),
        (_('Permissions'), {'fields': ('is_active', 'is_staff', 'is_superuser')}),
        (_('Important dates'), {'fields': ('last_login', 'date_joined')}),
    )


class Image(models.Model):
    user = models.ForeignKey(User, blank=True, null=True)
    name = models.CharField(max_length=100, unique=True)
    repo_addr = models.CharField(max_length=250)
    desc = models.CharField(max_length=1024, blank=True)
    create_time = models.DateTimeField()
    public = models.BooleanField()

    def __str__(self):
        return self.name


class Docker(models.Model):
    name = models.CharField(max_length=30)
    pod_name = models.CharField(max_length=100)
    service_name = models.CharField(max_length=100)
    app_port = models.IntegerField(default=8888)
    status = models.CharField(max_length=30)
    user = models.ForeignKey(User)
    cpu = models.IntegerField()
    gpu_mode = models.IntegerField(default=1)
    gpu = models.IntegerField()
    mem = models.IntegerField()
    password = models.CharField(max_length=50, blank=True)
    token = models.CharField(max_length=50, blank=True)
    image = models.ForeignKey(Image, blank=True, null=True, on_delete=models.SET_NULL)
    host_ip = models.GenericIPAddressField(blank=True, null=True)
    pod_ip = models.GenericIPAddressField(blank=True, null=True)
    last_used = models.DateTimeField(blank=True, null=True)
    start_time = models.DateTimeField()
    end_time = models.DateTimeField(blank=True, null=True)

    def __str__(self):
        return self.name


class Profile(models.Model):
    user = models.OneToOneField(User)
    max_docker = models.IntegerField(default=2,
                                     validators=[MaxValueValidator(200), MinValueValidator(1)])
    max_memory = models.IntegerField(default=4096,
                                     validators=[MaxValueValidator(81920), MinValueValidator(1024)],
                                     verbose_name="Docker memory limit (MB)")
    max_cpu = models.IntegerField(default=2,
                                  validators=[MaxValueValidator(32), MinValueValidator(1)],
                                  verbose_name="Docker CPU cores limit")
    max_gpu = models.IntegerField(default=0,
                                  validators=[MaxValueValidator(4), MinValueValidator(0)],
                                  verbose_name="Docker GPU cards limit")
    last_used_docker = models.ForeignKey(Docker, blank=True, null=True, on_delete=models.SET_NULL)

    def __str__(self):
        return self.user.__str__()


class Server(models.Model):
    MODE = (
        (1, 'Dedicated'),
        (0, 'Shared')
    )
    server_name = models.CharField(max_length=50)
    resource_mode = models.IntegerField(choices=MODE, default=1)

    def __str__(self):
        return self.server_name


@receiver(post_save, sender=Server)
def change_server_mode(sender, instance, created, **kwargs):
    try:
        if instance.resource_mode == 0:
            k8s.attach_node_label(instance.server_name, "gpu-usage-mode", "shared")
        else:
            k8s.attach_node_label(instance.server_name, "gpu-usage-mode", None)
    except:
        logger.error(traceback.format_exc())


@receiver(post_save, sender=User)
def create_user_profile(sender, instance, created, **kwargs):
    if created:
        try:
            os.mkdir(os.path.join(settings.NFS_PATH, instance.username))
            os.symlink('../public', os.path.join(settings.NFS_PATH, instance.username, 'public'))
        except:
            logger.error(traceback.format_exc())
        Profile.objects.create(
            user=instance
        )



@receiver(post_save, sender=User)
def save_user_profile(sender, instance, **kwargs):
    instance.profile.save()


def delete_user_dockers(user):
    try:
        ret = k8s.list_tensorflowapps_of_user(user.username)
    except Exception as e:
        logger.error(traceback.format_exc())
    if ret['status']:
        for app in ret['apps']:
            try:
                k8s.destroy_tensorflowapp(app['pod_name'], app['service_name'])
            except:
                logger.error(traceback.format_exc())


@receiver(pre_delete, sender=User)
def delete_user_profile(sender, instance, **kwargs):
    try:
        delete_user_dockers(instance)
        shutil.rmtree(os.path.join(settings.NFS_PATH, instance.username))
    except:
        logger.error(traceback.format_exc())
