##
# @file k8s.py
# @brief Interact with Kubernetes API server.
# @author Min(Jason) WANG - mwang14@lenovo.com
# @version 1.0
# @date 2016-12-20
from __future__ import print_function
from .config import config
import requests
import jinja2
import time
import random
import os
import json

env = jinja2.Environment(loader=jinja2.FileSystemLoader(os.path.dirname(
        __file__)))


def create_tensorflowapp(user_name, user_password=None,
                         mount_public_data=True,
                         cpu=1, mem=1024, gpu=0, mode=1,
                         token=None, image_tag=None,
                         applist=None):
    """
    Create a tensorflow pod and start the service so that users can access
    it through app_url.

    Args:
        user_name: the unique user name
        user_password: the user password to log into the tensorflow app
        mount_public_data: whether mount public data directory
        cpu: number of cpus requested, minimum 1
        mem: the requested memory size in (MB)
        gpu: the number of gpu
        mode: 0 stands for sharing; 1 stands for exclusive usage of GPU

    Returns:
        {'status': True if creation succeeds; False, otherwise,
         'pod_name': the name of the newly created pod,
         'service_name': the name of the newly created service,
         'app_url': the url that users can access the tensorflow app
        }
    """
    # first, create a pod
    res = {}
    user_data_dir = os.path.join(config['default'].NFS_ROOT,
                                 user_name)
    if mount_public_data:
        public_data_dir = os.path.join(config['default'].NFS_ROOT,
                                       config['default'].PUBLIC_DATA_DIR)
    else:
        public_data_dir = None

    _res = _create_a_pod(user_name, user_password, user_data_dir,
                         public_data_dir=public_data_dir,
                         cpu=cpu, mem=mem, gpu=gpu, mode=mode,
                         token=token, image_tag=image_tag,
                         applist=applist)
    if _res['status'] not in [requests.codes.ok, requests.codes.created]:
        res['status'] = False
        return res
    res['pod_name'] = _res['pod_name']
    # second, create a service
    _res2 = _create_a_service(user_name, _res['pod_name'])
    if _res2['status'] not in [requests.codes.ok, requests.codes.created]:
        res['status'] = False
        return res
    res['service_name'] = _res2['service_name']
    res['app_port'] = _res2['node_port']
    res['status'] = True
    return res


def destroy_tensorflowapp(pod_name, service_name):
    """
    Destroy the pod and service object of the tensorflow app.

    Params:
        pod_name: the name of the pod
        service_name: the name of the service

    Returns:
        {'status': true if succeeds; false, otherwise}
    """
    res = {}
    _res = _destroy_a_service(service_name)
    if _res['status'] not in [requests.codes.ok, requests.codes.not_found]:
        res['status'] = False
        return res
    _res2 = _destroy_a_pod(pod_name)
    if _res2['status'] not in [requests.codes.ok, requests.codes.not_found]:
        res['status'] = False
        return res
    res['status'] = True
    return res


def list_tensorflowapps_of_user(user_name):
    """
    List the tensorflow apps owned by the specified user.

    Args:
        user_name: the user name

    Returns:
        {'status': true if succeeds; false, otherwise,
         'apps': [
            {'pod_name': xxx,
             'service_name': xxx,
             'app_url': xxx,
             'app_status': xxx,
             'service_status': 'Success' | 'Fail' }, ...
            ]
        }
    """
    res = {}
    _res = _list_pods_of_user(user_name)
    _res2 = _list_services_of_user(user_name)
    if _res['status'] == requests.codes.ok and \
                    _res2['status'] == requests.codes.ok:
        res['status'] = True
    else:
        res['status'] = False
        return res
    apps = []
    for pod in _res['pods']:
        _d = {}
        _d['pod_name'] = pod['pod_name']
        _d['app_status'] = pod['pod_status']
        _d['host_ip'] = pod['host_ip']
        _d['pod_ip'] = pod['pod_ip']
        service_name = __get_servicename(pod['pod_name'])
        svs = [sv for sv in _res2['services']
               if sv['service_name'] == service_name]
        if len(svs) == 0:
            _d['service_status'] = 'Fail'
            _d['service_name'] = ''
        else:
            _d['service_status'] = 'Success'
            _d['service_name'] = svs[0]['service_name']
        apps.append(_d)
    res['apps'] = apps
    return res


def _list_pods_of_user(user_name):
    """
    List the tensorflow pods owned by the specified user.

    Args:
        user_name: the user name

    Returns:
        {'status': 200,
         'pods': [
            {'pod_name': xxx,
             'start_time': xxx,
             'pod_status': xxx,
             'host_ip': xxx
            ]
        }
    """
    headers = {'Authorization': config['default'].BEARER_TOKEN}
    url = '{0}/namespaces/default/pods'.format(
            config['default'].K8S_API_SERVER)
    params = {'labelSelector': 'username in ({0})'.format(user_name)}
    response = requests.get(url=url, headers=headers, params=params)
    res = {}
    res['status'] = response.status_code
    res['pods'] = []
    if response.status_code == requests.codes.ok:
        resp_json = response.json()
        for pod in resp_json['items']:
            _d = {}
            _d['pod_name'] = pod['metadata']['name']
            _d['pod_status'] = pod['status']['phase']
            _d['start_time'] = pod['status'].get('startTime', '')
            _d['host_ip'] = pod['status'].get('hostIP', '')
            _d['pod_ip'] = pod['status'].get('podIP', '')
            res['pods'].append(_d)
    response.close()
    return res


def _list_services_of_user(user_name):
    """
    List the tensorflow services owned by the specified user.
    Args:
        user_name: the user name

    Returns:
        {'status': 200,
         'services': [
            {'service_name': xxx,
             'cluster_ip': xxx,
             'target_port': xxx,
             'type': xxx,
             'node_port': xxx
            }
         ]}
    """
    headers = {'Authorization': config['default'].BEARER_TOKEN}
    url = '{0}/services'.format(config['default'].K8S_API_SERVER)
    params = {'labelSelector': 'username in ({0})'.format(user_name)}
    response = requests.get(url=url, headers=headers, params=params)
    res = {}
    res['status'] = response.status_code
    res['services'] = []
    if response.status_code == requests.codes.ok:
        resp_json = response.json()
        for sv in resp_json['items']:
            _d = {}
            _d['service_name'] = sv['metadata']['name']
            _d['cluster_ip'] = sv['spec']['clusterIP']
            _d['type'] = sv['spec']['type']
            _d['target_port'] = sv['spec']['ports'][0]['targetPort']
            _d['node_port'] = sv['spec']['ports'][0]['nodePort']
            res['services'].append(_d)
    return res


def _destroy_a_pod(pod_name):
    """
    Destroy a pod specified by name.

    Args:
        pod_name: the pod name

    Returns:
        {'status': 200 | 404}
    """
    headers = {'Authorization': config['default'].BEARER_TOKEN}
    url = '{0}/namespaces/default/pods/{1}'.format(
            config['default'].K8S_API_SERVER, pod_name)
    response = requests.delete(url=url, headers=headers)
    res = {}
    res['status'] = response.status_code
    return res


def _create_a_pod(user_name, user_password, user_data_dir,
                  public_data_dir=None, cpu=1, mem=1024, gpu=0,
                  mode=1, token=None, image_tag=None, applist=None):
    """
    Create a new tensorflow pod for the user.

    The new pod name will be: deepnex-<user_name>-<time>

    Note that: by convention, the names of Kubernetes resources should be up to
    maximum length of 253 characters and consists of lower case alphanumeric
    characters, -, and ., but certain resources have more specific
    restrictions.

    Args:
        user_name: the user name
        user_password: the password to log into the tensorflow app
        user_data_dir: the local dir for user to store
        public_data_dir: the public dir that is shared by all users
        cpu: the requested cpu numbers
        mem: the requested mem size in MB
        gpu: the requested number of gpus
        mode: 0 stands for shared; 1 stands for exclusive usage of GPUs

    Returns:
        {'status': 201,
         'pod_name': xxx,
         'creation_timestamp': xxx,
         'state': xxx}

    """
    headers = {'Authorization': config['default'].BEARER_TOKEN,
               'Content-Type': 'application/json'
               }
    url = '{0}/namespaces/default/pods/'.format(
            config['default'].K8S_API_SERVER)
    pod_name = 'deepnex-{0}-{1}'.format(user_name,
                                        int(time.mktime(time.localtime())))
    params = {'pod_name': pod_name, 'user_name': user_name, 'mem': mem,
              'cpu': cpu, 'gpu': gpu, 'user_data_dir': user_data_dir,
              'web_host_endpoint': config['default'].WEB_HOST_ENDPOINT,
              'nvidia_driver': config['default'].NVIDIA_DRIVER,
              'image_tag': "yaoman3/deeplearning:latest",
              'applist': applist
              }
    if public_data_dir is not None:
        params['public_data_dir'] = public_data_dir
    if user_password is not None:
        params['user_password'] = user_password
    if image_tag is not None:
        params['image_tag'] = image_tag
    if token is not None:
        params['token'] = token
    if mode == 0: # shared mode
        template = env.get_template(config['default'].DEEPNEX_POD_SHARED_JSON)
        gpu = min(gpu, config['default'].MAX_GPU_NUM)
        devs = list(range(0, config['default'].MAX_GPU_NUM))
        random.shuffle(devs)
        gpu_devices = ",".join(map(str, sorted(devs[0:gpu])))
        params['gpu_devices'] = gpu_devices
    else:
        template = env.get_template(
            config['default'].DEEPNEX_POD_EXCLUSIVE_JSON)
    payload = template.render(params=params)
    response = requests.post(url=url, headers=headers, data=payload)
    res = {}
    res['status'] = response.status_code
    if response.status_code in [requests.codes.ok, requests.codes.created]:
        resp_json = response.json()
        res['pod_name'] = resp_json['metadata']['name']
        res['creation_timestamp'] = resp_json['metadata']['creationTimestamp']
        res['state'] = resp_json['status']['phase']
    return res


def _destroy_a_service(service_name):
    """
    Destroy a service specified by name.

    Args:
        service_name: the kubernetes service name

    Returns:
        {'status': 200 | 404,
         'state': 'Success'}
    """
    headers = {'Authorization': config['default'].BEARER_TOKEN}
    url = '{0}/namespaces/default/services/{1}'.format(
            config['default'].K8S_API_SERVER, service_name)
    response = requests.delete(url=url, headers=headers)
    res = {}
    res['status'] = response.status_code
    if response.status_code == requests.codes.ok:
        resp_json = response.json()
        res['state'] = resp_json.get('status', 'Unknown')
    return res


def _create_a_service(user_name, pod_name):
    """
    Create a kubernetes service for the specified pod.

    The purpose of service is to open an access point so that external users
    can access the tensorflow pod.

    Args:
        user_name: the user name
        pod_name: the corresponding pod name

    Returns:
        {'status': 201,
         'service_name': xxx,
         'cluster_ip': xxx,
         'type': xxx,
         'target_port': xxx,
         'node_port': xxx}
    """
    headers = {'Authorization': config['default'].BEARER_TOKEN,
               'Content-Type': 'application/json'
               }
    url = '{0}/namespaces/default/services/'.format(
            config['default'].K8S_API_SERVER)
    service_name = __get_servicename(pod_name)
    params = {'user_name': user_name,
              'service_name': service_name,
              'pod_name': pod_name
              }
    template = env.get_template(config['default'].DEEPNEX_SERVICE_JSON)
    payload = template.render(params=params)
    response = requests.post(url=url, headers=headers, data=payload)
    res = {}
    res['status'] = response.status_code
    if response.status_code in [requests.codes.ok, requests.codes.created]:
        resp_json = response.json()
        res['service_name'] = resp_json['metadata']['name']
        res['cluster_ip'] = resp_json['spec']['clusterIP']
        res['type'] = resp_json['spec']['type']
        res['target_port'] = resp_json['spec']['ports'][0]['targetPort']
        res['node_port'] = resp_json['spec']['ports'][0]['nodePort']
    return res


def __get_servicename(pod_name):
    """
    Build the service name from pod_name.

    For example:
        if pod_name is: deepnex-mwang14-1234525, then
        the service_name is: deepnexservice-mwang14-1234525.

    Args:
        pod_name: the pod name

    Returns:
        The service name
    """
    service_name = 'deepnexservice-{0}'.format(
            pod_name[pod_name.index('-') + 1:])
    return service_name


def __get_public_endpoint(node_port):
    """
    Pick up an access server in Round-Robin manner.

    Returns:
        The public access endpoint.
    """
    index = random.randint(0,
                           len(config['default'].PUBLIC_ACCESS_ENDPOINTS) - 1)
    public_url = 'http://{0}:{1}'.format(
            config['default'].PUBLIC_ACCESS_ENDPOINTS[index], node_port)
    return public_url


def attach_node_label(node_name, key, value):
    """
    Attach the label key=value to the node.

    Note: When the value is None, the label is going to be removed.

    Args:
        node_name: the node name
        key: the label name
        value: the label value

    Returns:
        {'status': 200}
    """
    res = {}
    if key is None or len(key) == 0 or \
       node_name is None or len(node_name) == 0:
        res['status'] = 404
        return res
    headers = {'Authorization': config['default'].BEARER_TOKEN,
               'Content-Type': 'application/merge-patch+json'}
    url = '{0}/nodes/{1}'.format(config['default'].K8S_API_SERVER, node_name)
    _d = {'metadata': {'labels': { } } }
    _d['metadata']['labels'][key] = value
    payload = json.dumps(_d, indent=2)
    response = requests.patch(url=url, headers=headers, data=payload)
    res['status'] = response.status_code
    return res


def get_system_services():
    headers = {'Authorization': config['default'].BEARER_TOKEN,
               'Content-Type': 'application/json'
               }
    url = '{0}/namespaces/kube-system/services/'.format(
        config['default'].K8S_API_SERVER)
    response = requests.get(url=url, headers=headers)
    return response.json()


def get_nodes_status():
    headers = {'Authorization': config['default'].BEARER_TOKEN,
               'Content-Type': 'application/json'
               }
    url = '{0}/nodes/'.format(
        config['default'].K8S_API_SERVER)
    response = requests.get(url=url, headers=headers)
    return response.json()


if __name__ == '__main__':
    """res = attach_node_label('deepnex01', "min_label", "min_value")
    print(res)"""
    res = create_tensorflowapp('min', user_password='123', gpu=0, mode=1)
    print(res)
    res = list_tensorflowapps_of_user('min')
    print(res)
    """for app in res['apps']:
        if app['service_status'] == 'Success' and \
                        app['app_status'] == 'Running':
            res2 = destroy_tensorflowapp(app['pod_name'], app['service_name'])
            print(res2)"""
