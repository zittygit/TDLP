from __future__ import print_function
import jinja2
import os
from config import config

env = jinja2.Environment(
        loader=jinja2.FileSystemLoader(
                [os.path.dirname(__file__),
                 os.path.join(os.path.dirname(__file__),
                              'nfs_control')]))


def test_deepnex_pod_json():
    params = {}
    params['pod_name'] = 'tensorflow'
    params['user_name'] = 'min'
    params['user_password'] = '123456'
    params['web_host_endpoint'] = 'http://10.127.1.58:8000/'
    params['cpu'] = 1
    params['mem'] = 1024
    params['gpu'] = 1
    params['mode'] = 0
    params['user_data_dir'] = '/tmp'
    params['public_data_dir'] = '/tmp'
    template = env.get_template(config['default'].DEEPNEX_POD_SHARED_JSON)
    with open('temp/deepnex_pod_shared.json', 'w') as output:
        output.write(template.render(params=params))


if __name__ == '__main__':
    test_deepnex_pod_json()
