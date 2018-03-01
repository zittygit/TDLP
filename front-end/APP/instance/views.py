#! /usr/bin/env python
# -*- coding: utf-8 -*-
from flask import render_template, redirect, request, url_for, current_app, abort, jsonify, session
from flask_login import login_user, logout_user, current_user, login_required
from ..util.authorize import user_auth
from ..util.API_manage import ApiClient
from ..util.utils import get_login_data, get_secure_api_client
from ..models import Application
from . import instance
import json


@instance.route('/', methods=['GET', 'POST'])
@login_required
@user_auth
def index():
    api_client = get_secure_api_client(session, current_user)
    data = api_client.get_instances_list()['instances']

    # 由于管理员用户获取的实例是系统所有实例，因此该句的意义在于使管理员只获取自己的实例
    if current_user.permissions == 0:
        pure_data = []
        for i in data:
            if i['username'] == current_user.username:
                pure_data.append(i)
        data = pure_data

    return render_template("instance/index.html",
                           instances=data,
                           title=u"我的实例")


@instance.route('/<int:iid>', methods=['GET', 'POST'])
@login_required
@user_auth
def detail(iid):
    api_client = get_secure_api_client(session, current_user)
    i_data = api_client.get_instance_detail(iid=iid)
    instance = i_data['instance']
    config = i_data['config'][0]
    param = json.loads(i_data['config'][0]['param'])
    proxy = api_client.get_instance_proxy(iid=iid)

    return render_template("instance/detail.html",
                           instance=instance,
                           config=config,
                           param=param,
                           proxys=proxy['Services'],
                           title=u"实例详情")


@instance.route('/cookie', methods=['GET', 'POST'])
@login_required
@user_auth
def get_cookie():
    api_client = ApiClient(login_data=get_login_data(current_user))
    if api_client.login():
        return jsonify(api_client.get_login_cookie())
    else:
        return jsonify(message='login failed!')


@instance.route('/create/<int:app_id>', methods=['POST'])
@login_required
@user_auth
def create(app_id):
    app = Application.query.filter_by(aid=app_id).first_or_404()

    name = request.form['name']
    cpu = int(request.values.get("cpu"))
    memory = int(request.values.get("memory"))
    param_dict = dict(cpu=cpu, memory=memory)
    param_list = json.loads(app.param)
    for p in param_list:
        if p['type'] == 'varchar':
            param_dict[p['name']] = request.form[p['name']]
        else:
            param_dict[p['name']] = int(request.form[p['name']])
    param = json.dumps(param_dict)

    api_client = get_secure_api_client(session, current_user)
    print api_client.create_instance(app_id, name, param)
    return redirect(url_for('instance.index'))


@instance.route('/delete', methods=['POST'])
@login_required
@user_auth
def delete():
    iid = request.form['iid']
    api_client = get_secure_api_client(session, current_user)
    print api_client.delete_instance(iid)
    return jsonify(status='ok')
