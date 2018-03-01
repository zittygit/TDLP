#! /usr/bin/env python
# -*- coding: utf-8 -*-

import requests
import json


DLP410_API_URL = 'http://10.127.48.18:8080/'    # 超算410的API环境
DLP523_API_URL = 'http://172.18.232.157:8080/'  # 超算523的API环境
DLP103_API_URL = 'http://222.200.180.221:8080/' # B103的API环境
BASE_URL = DLP103_API_URL  # 选择取哪一个API环境

# 由于后台提供的过期时间无法获取到，因此需要自己设置过期时间，该值必须小于真实过期时长
TOKEN_EXPIRE_HOUR = 20  # 登录token的有效时长，单位为小时。
TIME_OUT = 3            # 请求默认允许响应时间，单位是秒


class ApiClient:
    def __init__(self, base_url=BASE_URL, login_cookie=None, login_data=None):
        self.base_url = base_url
        self.login_cookie = login_cookie
        self.login_data = login_data

    def login(self):
        """
        用户登录，返回值为是否成功（布尔型）
        """
        if self.login_data is None:
            print 'There is no login data'
            return False
        else:
            login_url = self.base_url + 'auth'
            r = requests.post(login_url, data=self.login_data)
            if r.status_code == 200:
                self.login_cookie = dict(kubernetes_token=r.cookies['kubernetes_token'])
                print 'login success'
                return True
            else:
                print 'login failed:' + r.text
                return False

    def get_login_cookie(self):
        """
        请求API_Server,获取登录cookie
        """
        login_url = self.base_url + 'auth'
        r = requests.post(login_url, data=self.login_data)
        if r.status_code == 200:
            return dict(kubernetes_token=r.cookies['kubernetes_token'])
        else:
            return None

    def open(self, url, data=None, method=None):
        """ 
        收发请求的底层函数，执行此函数前需要手动登录。
        将接收的字典格式数据data封装进请求，根据url参数发送到对应的api接口的地址，
        并对返回的数据进行处理以及保存。
        返回值为一个字典，即api接口返回的数据。
        """
        url = self.base_url + url
        if method == 'POST' or method == 'post':
            r = requests.post(url, data=data, cookies=self.login_cookie, timeout=TIME_OUT)
            return dict(data=r.json(), status=r.status_code)
        elif method == 'GET' or method == 'get':
            r = requests.get(url, params=data, cookies=self.login_cookie, timeout=TIME_OUT)
            return dict(data=r.json(), status=r.status_code)
        elif method == 'DELETE' or method == 'delete':
            r = requests.delete(url, data=data, cookies=self.login_cookie, timeout=TIME_OUT)
            return dict(data=r.json(), status=r.status_code)
        else:
            return dict(message='method is invalid', status=400)

    """
        用户管理
    """
    def get_user(self):
        """
        获取后台用户列表(仅管理员可用)
        """
        return self.open('user', method='get')['data']

    def create_user(self, username, password):
        """
        创建后台用户(仅管理员可用)
        """
        data = '{"username":"%s","password":"%s"}' % (username, password)
        return self.open('user', method='post', data=data)['status']

    """
        模板管理
    """
    def create_app(self, appname, path, info, param):
        """

        """
        data = '{"appname":"%s","path":"%s","info":"%s","param":"%s"}' % (appname, path, info, param)
        return self.open('app', method='post', data=data)['data']

    def delete_app(self, aid):
        """
        删除一个模板(仅管理员可用)
        """
        data = '{"aid":%d}' % aid
        return self.open('app', method='delete', data=data)['status']

    """
        实例管理
    """
    def create_instance(self, app_id, name, param):
        """
        创建一个实例 
        """
        data = json.dumps(dict(instancename=name, aid=app_id, param=param))
        print data
        return self.open('instance', method='post', data=data)

    def delete_instance(self, iid):
        """
        删除一个实例
        """
        data = '{"iid":%s}' % iid
        return self.open('instance', method='delete', data=data)['status']

    def get_instances_list(self):
        """
        获取实例列表,参数kind为必须(all、single、proxy)
        """
        data = dict(kind='all')
        return self.open('instance', method='get', data=data)['data']

    def get_instance_detail(self, iid):
        """
        获取某个实例的详细信息，包括配置参数
        """
        data = dict(kind='single', iid=iid)
        return self.open('instance', method='get', data=data)['data']

    def get_instance_proxy(self, iid):
        """
        获取某个实例的proxy代理链接
        """
        data = dict(kind='proxy', iid=iid)
        return self.open('instance', method='get', data=data)['data']


if __name__ == '__main__':
    ADMIN_LOGIN_DATA = '{"username":"long","password":"123456"}'
    api_client = ApiClient(login_data=ADMIN_LOGIN_DATA)
    api_client.login()
