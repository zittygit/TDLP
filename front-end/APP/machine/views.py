#! /usr/bin/env python
# -*- coding: utf-8 -*-
from . import machine
from flask import render_template, redirect, request, url_for, current_app, abort, jsonify, session
from flask_login import login_required, current_user
from ..util.utils import get_secure_api_client
from .. import db
import datetime
import json


@machine.route('/bill')
@login_required
def bill():
    """显示用户的机时统计首页"""
    api_client = get_secure_api_client(session, current_user)
    instance_list = api_client.get_instances_list()['instances']
    bill_dict = dict()
    for i in instance_list:
        if i['appname'] not in bill_dict:
            bill_dict[i['appname']] = compute_hours(i)
        else:
            bill_dict[i['appname']] = bill_dict[i['appname']] + compute_hours(i)

    return render_template('machine/bill.html',
                           title=u'机时账单',
                           bill_dict=bill_dict)


@machine.route('/detail/<string:app_name>')
@login_required
def detail(app_name):
    """单个应用模板的实例机时分布"""
    api_client = get_secure_api_client(session, current_user)
    instance_list = api_client.get_instances_list()['instances']
    app_list = []
    for i in instance_list:
        if i['appname'] == app_name:
            i_dict = dict()
            i_dict['name'] = i['instancename']
            i_dict['hours'] = compute_hours(i)
            i_dict['create_time'] = i['createtime']
            i_dict['state'] = i['state']
            app_list.append(i_dict)
    return render_template('machine/detail.html',
                           title=u'机时详情',
                           app_list=app_list,
                           app_name=app_name)


@machine.route('/resources')
@login_required
def resources():
    """用户的资源配额"""
    api_client = get_secure_api_client(session, current_user)
    raw_instances = api_client.get_instances_list()['instances']
    instances = []
    for i in raw_instances:
        if i['state'] == 0 and i['username'] == current_user.username:
            instances.append(i)

    for i in instances:
        i_data = api_client.get_instance_detail(iid=i['iid'])
        param = json.loads(i_data['config'][0]['param'])
        i['cpu'] = param['cpu']
        i['memory'] = param['memory']

    return render_template('machine/resources_index.html',
                           instances=instances,
                           title=u'资源管理')


def compute_hours(instance):
    """bill()的工具函数，用于计算一个实例的所用机时"""
    if instance["deletetime"] != "":
        start = datetime.datetime.strptime(instance["createtime"], '%Y-%m-%d %H:%M:%S')
        end = datetime.datetime.strptime(instance["deletetime"], '%Y-%m-%d %H:%M:%S')
        return (end.day - start.day) * 24 + end.hour - start.hour
    else:
        start = datetime.datetime.strptime(instance["createtime"], '%Y-%m-%d %H:%M:%S')
        now = datetime.datetime.now()
        return (now.day - start.day) * 24 + now.hour - start.hour


@machine.route('/bill/app_chart')
@login_required
def bill_app_chart():
    """用于机时统计页的应用图表——应用机时消耗图"""
    api_client = get_secure_api_client(session, current_user)
    instance_list = api_client.get_instances_list()['instances']
    bill_dict = dict()
    for i in instance_list:
        if i['appname'] not in bill_dict:
            bill_dict[i['appname']] = compute_hours(i)
        else:
            bill_dict[i['appname']] += compute_hours(i)
    return jsonify(apps=bill_dict.keys(), hours=bill_dict.values())


@machine.route('/bill/ins_chart')
@login_required
def bill_ins_chart():
    """用于机时统计页的应用图表——实例计时消耗图"""
    api_client = get_secure_api_client(session, current_user)
    instance_list = api_client.get_instances_list()['instances']
    bill_dict = dict()
    for i in instance_list:
        if i['state'] == 0:
            bill_dict[i['instancename']] = compute_hours(i)
        else:
            if 'deleted' not in bill_dict.keys():
                bill_dict['deleted'] = compute_hours(i)
            else:
                bill_dict['deleted'] += compute_hours(i)
    return jsonify(inss=bill_dict.keys(), hours=bill_dict.values())


@machine.route('/detail/dist_chart/<string:app_name>')
@login_required
def detail_dist_chart(app_name):
    """用于机时详情页的图表——实例机时分布图"""
    api_client = get_secure_api_client(session, current_user)
    instance_list = api_client.get_instances_list()['instances']
    the_dict = dict()
    the_dict['deleted'] = 0
    for i in instance_list:
        if i['appname'] == app_name:
            if i['state'] == 0:
                the_dict[i['instancename']] = compute_hours(i)
            else:
                the_dict['deleted'] += compute_hours(i)

    if the_dict['deleted'] == 0:
        del the_dict['deleted']
        return jsonify(inss=the_dict.keys(), hours=the_dict.values())
    else:
        deleted = the_dict['deleted']
        del the_dict['deleted']
        inss = the_dict.keys()
        hours = the_dict.values()
        inss.append(u"已删除实例")
        hours.append(deleted)
        return jsonify(inss=inss, hours=hours)
