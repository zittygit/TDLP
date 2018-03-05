from django.utils import timezone
import os
import secrets
import time
from django.shortcuts import render, redirect
from django.contrib.auth.decorators import login_required
from django.core import serializers
from django.http import HttpResponse,JsonResponse
from django.conf import settings
from django.db.models import Sum
import logging
import traceback
import re
import requests
import urllib
from sshtunnel import SSHTunnelForwarder
from .models import Docker, Profile, User, Image, Server
from k8s_api import k8s
from k8s_api.config import Config as cf

# Create your views here.


logger = logging.getLogger('servers')
tunnels = dict()
last_request_time = 0
service_grafana = None
service_dashboard = None
cpu_resources = 0
mem_resources = 0
gpu_resources = 0
nodes_list = list()


def update_node_data():
    global cpu_resources
    global mem_resources
    global gpu_resources
    global nodes_list
    try:
        ret = k8s.get_nodes_status()
        nodes = ret.get('items')
        servers = Server.objects.all()
        node_modes = dict()
        if nodes:
            for node in nodes:
                cpu_resources += int(node['status']['capacity']['cpu'])
                mem_resources += int(int(node['status']['capacity']['memory'][:-2])/1024)
                gpu_resources += int(node['status']['capacity']['alpha.kubernetes.io/nvidia-gpu'])
                for addr in node['status']['addresses']:
                    if addr['type']=='LegacyHostIP':
                        nodes_list.append(addr['address'])
                if 'gpu-usage-mode' in node['metadata']['labels'] and \
                        node['metadata']['labels']['gpu-usage-mode']=='shared':
                    gpu_mode = 0
                else:
                    gpu_mode = 1
                node_modes[node['metadata']['name']] = gpu_mode
        servers = Server.objects.all()
        for server in servers:
            if server.server_name not in node_modes:
                server.delete()
            else:
                server.resource_mode = node_modes[server.server_name]
                server.save()
                node_modes[server.server_name] = None
        for node_name, node_mode in node_modes.items():
            if node_mode is not None:
                server = Server(server_name=node_name, resource_mode=node_mode)
                server.save()

        # Reserve resources for system
        cpu_resources -= 2
        mem_resources -= 2048
        return True
    except:
        logger.error(traceback.format_exc())
        return False


def create_tunnel(app):
    try:
        if app.pod_name not in tunnels:
            remote_server = re.findall(r'[0-9]+(?:\.[0-9]+){3}', cf.K8S_API_SERVER)[0]
            local_server = re.findall(r'[0-9]+(?:\.[0-9]+){3}', cf.WEB_HOST_ENDPOINT)[0]
            if remote_server == local_server:
                return None
            tunnels[app.pod_name] = SSHTunnelForwarder(
                remote_server,
                ssh_username=cf.K8S_SERVER_USERNAME,
                remote_bind_address=('127.0.0.1', app.app_port),
                local_bind_address=('0.0.0.0', app.app_port),
                set_keepalive=300
            )
        tunnels[app.pod_name].start()
        return tunnels[app.pod_name]
    except:
        logger.error(traceback.format_exc())
        return None


def remove_tunnel(app):
    if app.pod_name in tunnels:
        tunnels[app.pod_name].stop()
        del tunnels[app.pod_name]


def get_used_resource():
    capacity = {}
    usage = {}
    percent = {}
    try:
        if cpu_resources==0:
            update_node_data()
        apps = Docker.objects.exclude(status='deleted')
        stats = apps.aggregate(used_cpu=Sum("cpu"), used_mem=Sum("mem"), used_gpu=Sum("gpu"))
        for key, value in stats.items():
            stats[key] = 0 if value is None else value
            usage[key[-3:]] = stats[key]
        usage['mem'] /= 1024.0  # use GB as unit
        capacity['cpu'] = cpu_resources
        capacity['gpu'] = gpu_resources
        capacity['mem'] = float("%.1f" % (mem_resources / 1024.0))
        disk = os.statvfs(settings.NFS_PATH)
        usage['disk'] = float("%.1f" % (disk.f_bfree / 1024.0)) # use GB as unit
        capacity['disk'] = float("%.1f" % (disk.f_blocks / 1024.0)) # use GB as unit
        for key in capacity:
            percent[key] = float("%.1f" %(usage[key] * 100 / capacity[key]))
    except:
        logger.error(traceback.format_exc())
        for key in usage:
            usage[key] = 0
            percent[key] = 0
        usage['disk'] = 0
        capacity['disk'] = 0
        percent['disk'] = 0

    for key in usage:
        if usage[key] > capacity[key]:
            usage[key] = capacity[key]
            percent[key] = 100
    usage['disk'] = capacity['disk'] - usage['disk']
    percent['disk'] = 100 - percent['disk']

    strData = {}
    strData['disk_capacity'] = ("%.1f" % (capacity['disk']/1024.0) + 'T' ) if capacity['disk'] >= 1000 else (str(capacity['disk']) + 'G')
    strData['disk_usage'] = ("%.1f" % (usage['disk']/1024.0) + 'T' ) if usage['disk'] >= 1000 else (str(usage['disk']) + 'G')
    strData['mem_capacity'] = ("%.1f" % (capacity['mem']/1024.0) + 'T') if capacity['mem'] >= 1000 else (str(capacity['mem']) + 'G')
    strData['mem_usage'] = ("%.1f" % (usage['mem']/1024.0) + 'T') if usage['mem'] >= 1000 else (str(usage['mem']) + 'G')

    return {"usage": dict(usage), "percent": dict(percent), "capacity": dict(capacity), "strData": dict(strData)}


def get_allother_dockers(user, update=True):
    all_apps = None
    for profile in Profile.objects.all():
        if profile.user != user:
            apps = get_all_dockers(profile.user, update)
            if all_apps==None:
                all_apps = apps
            else:
                all_apps = all_apps | apps
    return all_apps

def get_all_dockers(user, update=True):
    if update:
        try:
            ret = k8s.list_tensorflowapps_of_user(user.username)
        except:
            logger.error(traceback.format_exc())
            return []
        db_apps = Docker.objects.filter(user=user).exclude(status='deleted')
        server_apps = {}
        if ret['status']:
            for app in ret['apps']:
                server_apps[app['pod_name']] = app
        for app in db_apps:
            if app.pod_name not in server_apps:
                app.status = 'deleted'
                remove_tunnel(app)
                app.save()
            else:
                server_apps[app.pod_name]['found'] = True
                app.status = server_apps[app.pod_name]['app_status']
                app.host_ip = server_apps[app.pod_name]['host_ip']
                app.pod_ip = server_apps[app.pod_name]['pod_ip']
                app.save()

    apps = Docker.objects.filter(user=user).exclude(status='deleted')

    if update:
        # Remove all not used apps in server.
        for pod_name, app in server_apps.items():
            if not app.get('found', False):
                try:
                    k8s.destroy_tensorflowapp(app['pod_name'], app['service_name'])
                except:
                    logger.error(traceback.format_exc())

    return apps


def get_images(user):
    try:
        public_images = Image.objects.exclude(user=user).filter(public=True)
        user_images = Image.objects.filter(user=user)
        return {'public_images': public_images, 'user_images': user_images}
    except:
        return {'public_images': None, 'user_images': None}


@login_required
def home(request):
    #apps = Docker.objects.filter(user=request.user)
    apps = get_all_dockers(request.user)
    profile = Profile.objects.get(user=request.user)
    appid = None
    if request.method=='GET':
        if 'appid' in request.GET:
            appid = request.GET['appid']
    elif request.method=='POST':
        appid = request.POST.get('appid', '')
    try:
        if appid:
            app = Docker.objects.exclude(status='deleted').get(pod_name=appid)
            profile = Profile.objects.get(user=request.user)
            profile.last_used_docker = app
            profile.save()
    except:
        logger.error(traceback.format_exc())
    try:
        if not profile.last_used_docker and len(apps)!=0:
            profile.last_used_docker = apps[0]
            profile.save()
    except:
        logger.error(traceback.format_exc())
        profile.last_used_docker = None
        profile.save()
    if profile.last_used_docker and profile.last_used_docker.status=='deleted':
        profile.last_used_docker = None
        profile.save()
    usage = get_used_resource()
    cpu = profile.max_cpu
    mem = profile.max_memory
    gpu = profile.max_gpu
    images = get_images(request.user)
    return render(request, 'home.html',
                  {'usage': usage,
                   'apps': apps,
                   'last_used': profile.last_used_docker,
                   'cpu': cpu, 'mem': mem, 'gpu': gpu, 'images': images,
                   'refresh': True})


@login_required
def workspace(request):
    app = None
    profile = Profile.objects.get(user=request.user)
    if 'appid' in request.GET:
        app = Docker.objects.exclude(status='deleted').get(pod_name=request.GET['appid'])
        profile = Profile.objects.get(user=request.user)
        profile.last_used_docker = app
        profile.save()
    else:
        if profile.last_used_docker and profile.last_used_docker.status!='deleted':
            app = profile.last_used_docker
    usage = get_used_resource()
    if app:
        http_host = request.META.get("HTTP_HOST", '')
        host_ip = http_host.split(':')[0]
        remote_server = re.findall(r'[0-9]+(?:\.[0-9]+){3}', cf.K8S_API_SERVER)[0]
        url = 'http://' + remote_server + ':' + str(app.app_port) + '/?token=' + app.token
        if cf.TUNNEL:
            url = 'http://'+host_ip+':'+str(app.app_port)+'/?token='+app.token
            create_tunnel(app)
    else:
        url = None
    apps = get_all_dockers(request.user)
    return render(request, 'workspace.html', {'usage': usage, 'app': app, 'apps': apps, 'app_url': url})


@login_required
def apps(request):
    user_profile = Profile.objects.get(user=request.user)
    cpu = user_profile.max_cpu
    mem = user_profile.max_memory
    gpu = user_profile.max_gpu
    usage = get_used_resource()
    images = get_images(request.user)
    return render(request, 'apps.html',
                  {'usage': usage, 'cpu': cpu, 'mem': mem, 'gpu': gpu, 'images': images})


@login_required
def app_info(request):
    apps = get_all_dockers(request.user)
    response = serializers.serialize("json", apps)
    return HttpResponse(response, content_type='application/json')


@login_required
def app_manage(request):
    username = request.user.username
    password = request.POST.get('token', '')
    gpu_mode = int(request.POST.get('gpu_mode', 1))
    gpu = int(request.POST.get('gpu', 0))
    mem = int(request.POST.get('mem', 1024))
    cpu = int(request.POST.get('cpu', 1))
    image_name = request.POST.get('image', None)
    appname = request.POST.get('appname', None)
    user_profile = Profile.objects.get(user=request.user)
    if image_name is not None:
        image = Image.objects.get(name=image_name)
        image_tag = image.repo_addr
    else:
        image = None
        image_tag = None

    if gpu>=3:
        gpu = 0
    if request.POST['action']=='Create':
        all_apps = get_all_dockers(request.user)
        if len(all_apps)>=user_profile.max_docker:
            message = 'You can only create {} apps!'.format(user_profile.max_docker)
            return JsonResponse({'status': False, 'message': message})
        for app in all_apps:
            if appname==app.name:
                message = 'Duplicated app name!'.format(user_profile.max_docker)
                return JsonResponse({'status': False, 'message': message})
        try:
            if not password:
                token = secrets.token_urlsafe(16)
            else:
                token = None
            applist = list()
            for app in all_apps:
                applist.append(app.pod_name)
            otherapps = get_allother_dockers(request.user)
            if otherapps:
                for app in otherapps:
                    applist.append(app.pod_name)
            if not os.path.isdir(os.path.join(settings.NFS_PATH, username)):
                os.mkdir(os.path.join(settings.NFS_PATH, username))
                os.symlink('../public', os.path.join(settings.NFS_PATH, username, 'public'))
            ret = k8s.create_tensorflowapp(
                user_name=username,
                cpu=cpu,
                gpu=gpu,
                mem=mem,
                user_password=password,
                token=token,
                image_tag=image_tag,
                applist=','.join(applist),
                mode=gpu_mode
            )
            if ret['status']:
                app = Docker(
                    name=appname,
                    pod_name=ret['pod_name'],
                    service_name=ret['service_name'],
                    app_port=ret['app_port'],
                    status='New',
                    user=request.user,
                    cpu=cpu,
                    gpu_mode=gpu_mode,
                    gpu=gpu,
                    mem=mem,
                    password=password,
                    token=token,
                    image=image,
                    start_time=timezone.localtime(timezone.now())
                )
                app.save()
            else:
                message = 'Create application failed!'
                return JsonResponse({'status': False, 'message': message})
        except:
            logger.error(traceback.format_exc())
            message = 'Internal Error!'
            return JsonResponse({'status': False, 'message': message})

    elif request.POST['action']=='Delete':
        app = Docker.objects.filter(user=request.user).exclude(status='deleted').get(name=appname)
        try:
            k8s.destroy_tensorflowapp(app.pod_name, app.service_name)
        except:
            logger.error(traceback.format_exc())
        app.status='deleted'
        app.end_time = timezone.localtime(timezone.now())
        app.save()
        remove_tunnel(app)
    return JsonResponse({'status': True})


@login_required
def profile(request):
    usage = get_used_resource()
    return render(request, 'profile.html', {'usage': usage})


@login_required
def user_profile(request):
    user = User.objects.get(username=request.user.username)
    password = request.POST.get('current_password', None)
    new_password = request.POST.get('new_password', None)
    first_name = request.POST.get('first_name', None)
    last_name = request.POST.get('last_name', None)
    email = request.POST.get('email', None)
    #status = False
    #usage = get_used_resource()
    if user.check_password(password):
        if new_password:
            user.set_password(new_password)
        user.first_name = first_name
        user.last_name = last_name
        user.email = email
        user.save()
        #status = True
        message = "Profile Updated!"
        return JsonResponse({'status': True, 'message': message})
    else:
        message = "Invalid Password!"
        return JsonResponse({'status': False, 'message': message})
    # return render(request, 'home.html', {'usage': usage,
    #                                         "status": status,
    #                                         "message": message} )


@login_required
def usage(request):
    usage = get_used_resource()
    nodes_names = list(range(1, len(nodes_list)+1))
    node_prefix = cf.NODE_PREFIX
    return render(request, 'usage.html', {'usage': usage, 'node_prefix': node_prefix,  'nodes_list': nodes_names})


@login_required
def images(request):
    images = get_images(request.user)
    usage = get_used_resource()
    return render(request, 'images.html',{'usage': usage, 'images':images})


@login_required
def image_manage(request):
    user = User.objects.get(username=request.user.username)
    name = request.POST.get('name', '')
    if request.POST['action'] == 'Add':
        try:
            repo_addr = request.POST.get('repo_addr', '')
            desc = request.POST.get('desc', '')
            public = request.POST.get('public', False)
            if public == 'True':
                public = True
            create_time = timezone.localtime(timezone.now())
            image = Image(user=user, name=name, repo_addr=repo_addr,
                          desc=desc, create_time=create_time, public=public)
            image.save()
        except:
            logger.error(traceback.format_exc())
            return JsonResponse({'status': "failed"})
        return JsonResponse({'status': "success"})
    elif request.POST['action'] == 'Update':
        repo_addr = request.POST.get('repo_addr', '')
        desc = request.POST.get('desc', '')
        public = request.POST.get('public', False)
        image = Image.objects.filter(user=user).get(name=name)
        image.repo_addr = repo_addr
        image.public = public
        image.desc = desc
        image.save()
        return JsonResponse({'status': "success"})
    elif request.POST['action'] == 'Delete':
        image = Image.objects.filter(user=user).get(name=name)
        image.delete()
        return JsonResponse({'status': "success"})


@login_required
def grafana_data(request, url=''):
    if service_grafana is None:
        update_service_info()
    param = urllib.parse.urlencode(request.GET)
    return redirect(urllib.parse.urljoin(service_grafana, url)+'?'+param)


@login_required
def dashboard_data(request, url=''):
    if service_dashboard is None:
        update_service_info()
    param = urllib.parse.urlencode(request.GET)
    if param:
        url = url+'?'+param
    ret = requests.get(urllib.parse.urljoin(service_dashboard, url)).json()
    return JsonResponse(ret)


def update_flag():
    global last_request_time
    update = False
    current_time = time.time()
    if current_time-last_request_time>4:
        update = True
        last_request_time = current_time
    return update


@login_required
def api_appdata(request, url=''):
    if service_dashboard is None:
        update_service_info()
    apps = get_all_dockers(request.user, update_flag())
    param = urllib.parse.urlencode(request.GET)
    if param:
        url = url + '?' + param
    ret = requests.get(urllib.parse.urljoin(service_dashboard, url)).json()
    for i in reversed(range(len(ret['pods']))):
        pod = ret['pods'][i]
        app = None
        try:
            if pod is not None and 'objectMeta' in pod and 'name' in pod["objectMeta"] \
                    and apps.filter(pod_name=pod["objectMeta"]["name"]).exists():
                app = apps.get(pod_name=pod["objectMeta"]["name"])
        except:
            logger.error(traceback.format_exc())
        if app:
            app_data = dict()
            app_data['name'] = app.name
            app_data['status'] = app.status
            app_data['cpu'] = app.cpu
            app_data['gpu'] = app.gpu
            app_data['mem'] = app.mem
            app_data['podIP'] = app.pod_ip
            app_data['gpu_mode']=app.gpu_mode
            if app.image:
                app_data['image'] = app.image.name
            else:
                app_data['image'] = None
            ret['pods'][i]['appMeta'] = app_data
        else:
            del ret['pods'][i]
    return JsonResponse(ret)


@login_required
def api_appdata_allothers(request, url=''):
    if service_dashboard is None:
        update_service_info()
    apps = get_allother_dockers(request.user, update_flag())
    param = urllib.parse.urlencode(request.GET)
    if param:
        url = url + '?' + param
    ret = requests.get(urllib.parse.urljoin(service_dashboard, url)).json()
    for i in reversed(range(len(ret['pods']))):
        pod = ret['pods'][i]
        app = None
        try:
            if pod is not None and 'objectMeta' in pod and 'name' in pod["objectMeta"] \
                    and apps.filter(pod_name=pod["objectMeta"]["name"]).exists():
                app = apps.get(pod_name=pod["objectMeta"]["name"])
        except:
            logger.error(traceback.format_exc())
        if app:
            app_data = dict()
            app_data['name'] = app.name
            app_data['status'] = app.status
            app_data['cpu'] = app.cpu
            app_data['gpu'] = app.gpu
            app_data['mem'] = app.mem
            app_data['podIP'] = app.pod_ip
            app_data['username'] = app.user.username
            app_data['gpu_mode'] = app.gpu_mode
            if app.image:
                app_data['image'] = app.image.name
            else:
                app_data['image'] = None
            ret['pods'][i]['appMeta'] = app_data
        else:
            del ret['pods'][i]
    return JsonResponse(ret)


def update_service_info():
    global service_grafana
    global service_dashboard
    try:
        k8s_host = re.findall(r'[0-9]+(?:\.[0-9]+){3}', cf.K8S_API_SERVER)[0]
        ret = k8s.get_system_services()
        for service in ret['items']:
            if service['metadata']['name']=='kubernetes-dashboard':
                service_dashboard = 'http://' + k8s_host + ':' + str(service['spec']['ports'][0]['nodePort'])
            if service['metadata']['name']=='monitoring-grafana':
                service_grafana = 'http://' + k8s_host + ':' + str(service['spec']['ports'][0]['nodePort'])
        return True
    except:
        logger.error(traceback.format_exc())
        return False
