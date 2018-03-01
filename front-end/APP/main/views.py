#! /usr/bin/env python
# -*- coding: utf-8 -*-
from . import main
from ..models import Application, News, Notice
from flask import render_template, jsonify, session
from flask_login import login_required, current_user
from ..util.utils import get_secure_api_client
from .. import db


@main.route('/')
@login_required
def index():
    # 计算用户资料完整度profile_ratio
    profile_num = 0
    if current_user.real_name is not None and current_user.real_name != '':
        profile_num += 1
    if current_user.phone is not None and current_user.phone != '':
        profile_num += 1
    if current_user.address is not None and current_user.address != '':
        profile_num += 1
    profile_ratio = 40 + profile_num * 20

    # 计算用户实例数量instance_num
    api_client = get_secure_api_client(session, current_user)
    instance_list = api_client.get_instances_list()['instances']
    instance_num = 0
    for i in instance_list:
        if i['state'] == 0 and i['username'] == current_user.username:
            instance_num += 1

    # 计算用户当前可用应用模板数量
    app_num = len(Application.query.all())

    # 取出最新的四条系统公告与四条新闻资讯
    news = News.query.order_by(db.desc(News.id)).limit(4)
    notices = Notice.query.order_by(db.desc(Notice.id)).limit(4)

    return render_template("main/index.html",
                           profile_ratio=profile_ratio,
                           instance_num=instance_num,
                           app_num=app_num,
                           news_list=news,
                           notice_list=notices,
                           title=u"我的主页")


@main.route('/app_dist')
@login_required
def load_app_distribute():
    """读取各应用实例的分布返回给前端ajax"""
    api_client = get_secure_api_client(session, current_user)
    instance_list = api_client.get_instances_list()['instances']
    app_dict = dict()
    for i in instance_list:
        if i['state'] == 0 and i['username'] == current_user.username:
            if i['appname'] not in app_dict:
                app_dict[i['appname']] = 1
            else:
                app_dict[i['appname']] += 1
    return jsonify(apps=app_dict.keys(), values=app_dict.values())


@main.app_errorhandler(404)
def page_404(err):
    return render_template('404.html', title='404'), 404


@main.app_errorhandler(403)
def page_403(err):
    return render_template('403.html', title='403'), 403


@main.app_errorhandler(500)
def page_500(err):
    return render_template('500.html', title='500'), 500
