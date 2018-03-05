class Config:
    BEARER_TOKEN = 'Bearer pSyPU+OB980aAFmvNIZwX7C+dMF2HTq6pDlU+TsNZvI='
    K8S_API_SERVER = 'http://10.127.1.153:8080/api/v1'
    DEEPNEX_POD_SHARED_JSON = 'deepnex_pod_shared.json.j2'
    DEEPNEX_POD_EXCLUSIVE_JSON = 'deepnex_pod_exclusive.json.j2'
    DEEPNEX_SERVICE_JSON = 'deepnex_service.json.j2'
    NODE_PREFIX = 'deepnex0'
    WEB_HOST_ENDPOINT = 'http://10.127.1.78:8000/'
    K8S_SERVER_USERNAME = 'bigdata'
    NFS_ROOT = '/mnt/nfs'
    PUBLIC_DATA_DIR = 'public'
    NVIDIA_DRIVER = '/var/lib/nvidia-docker/volumes/nvidia_driver/375.20/'
    MAX_GPU_NUM = 2
    TUNNEL = False


config = {
    'default': Config
}
