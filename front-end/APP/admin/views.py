#! /usr/bin/env python
# -*- coding: utf-8 -*-
from flask import render_template, redirect, request, url_for, current_app, abort, jsonify, session
from flask_login import login_user, logout_user, current_user
from datetime import datetime
import json
from ..models import User, News, Notice, Application
from ..util.authorize import admin_login
from ..util.file_manage import get_file_type
from ..util.API_manage import ApiClient
from PIL import Image
import os
from ..util.utils import get_login_data, get_secure_api_client
from . import admin
from .. import db


@admin.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'GET':
        return render_template('admin/admin_login.html', title=u'管理员登录')
    elif request.method == 'POST':
        _form = request.form
        u = User.query.filter_by(email=_form['email']).first()
        if u is None:
            message_e = u'邮箱不存在'
            return render_template('admin/admin_login.html',
                                   title=u"管理员登录",
                                   data=_form,
                                   message_e=message_e)
        if u and u.verify_password(_form['password']) and u.permissions == 0:
            login_user(u)
            u.last_login = datetime.now()
            db.session.commit()
            return redirect(url_for('admin.index'))
        else:
            message_p = u"密码错误,或该用户未注册为管理员"
            return render_template('admin/admin_login.html',
                                   title=u"管理员登录",
                                   data=_form,
                                   message_p=message_p)


@admin.route('/')
@admin_login
def index():
    """管理员的控制台主页"""
    return render_template('admin/index.html',
                           title=u'主控制台')


@admin.route('/user')
@admin_login
def user():
    """管理员查看系统用户(包括未通过认证的用户)列表"""
    users = User.query.all()
    return render_template('admin/user/index.html',
                           users=users,
                           title=u'用户管理')


@admin.route('/user/auth/<int:uid>', methods=['GET', 'POST'])
@admin_login
def user_auth(uid):
    """管理员审核通过普通用户的认证"""
    u = User.query.filter_by(id=uid).first_or_404()
    api_client = get_secure_api_client(session, current_user)
    if api_client.create_user(u.username, u.api_password) == 200:  # 200 为状态返回码
        u.is_auth = 1
        db.session.commit()
    return redirect(url_for('admin.user'))


@admin.route('/instance', methods=['GET', 'POST'])
@admin_login
def instance():
    """管理员查看系统所有实例"""
    api_client = get_secure_api_client(session, current_user)
    data = api_client.get_instances_list()
    return render_template("admin/instance/index.html",
                           instances=data['instances'],
                           title=u"所有实例")


@admin.route('/instance/<int:iid>', methods=['GET', 'POST'])
@admin_login
def instance_detail(iid):
    """管理员查看系统某个实例详情"""
    api_client = get_secure_api_client(session, current_user)
    i_data = api_client.get_instance_detail(iid=iid)
    instance = i_data['instance']
    config = i_data['config'][0]
    param = json.loads(config['param'])
    proxy = api_client.get_instance_proxy(iid=iid)
    return render_template("admin/instance/detail.html",
                           instance=instance,
                           config=config,
                           param=param,
                           proxys=proxy['Services'],
                           title=u"实例详情")


@admin.route('/application')
@admin_login
def application():
    """管理员查看系统所有应用模板"""
    apps = Application.query.all()
    return render_template("admin/application/index.html",
                           apps=apps,
                           title=u"模板管理")


@admin.route('/application/create', methods=['GET', 'POST'])
@admin_login
def application_create():
    """管理员创建新模板"""
    if request.method == 'GET':
        return render_template("admin/application/create.html",
                               title=u"模板创建")

    elif request.method == 'POST':
        # 把自定义预设参数的信息整理成json格式的字符串,在k8s端创建成功后写入服务器端数据库

        """保存在服务器端的参数字符串，相比于发送给k8s端的参数字符串，更多了一些用于向用户展示的说明信息"""
        param_num = int(request.values.get("param-select"))
        param = list()           # 保存在服务器端的参数字符串
        param_kub = dict()       # 发送给k8s端的参数字符串
        param_kub['cpu'] = 'int'
        param_kub['memory'] = 'int'
        for i in range(1, param_num+1):
            _name = "param_name_%d" % i
            _note = "param_note_%d" % i
            _type = "param_type_%d" % i
            param_name = request.form[_name]
            param_note = request.form[_note]
            param_type = request.form[_type]
            param.append(dict(name=param_name, note=param_note, type=param_type))
            param_kub[str(param_name)] = str(param_type)

        param_str = json.dumps(param_kub).replace('"', '\\"')
        api_client = get_secure_api_client(session, current_user)
        data = api_client.create_app(appname=request.form['name_en'],
                                     path=request.form['path'],
                                     info=request.form['info_en'],
                                     param=param_str)

        if 'aid' in data:  # 若成功返回结果
            new_app = Application(name=request.form['name_zh'],
                                  aid=data["aid"],
                                  info=request.form['info_zh'],
                                  path=request.form['path'],
                                  param=json.dumps(param),
                                  param_guide=request.form['param_guide'])
            db.session.add(new_app)
            db.session.commit()
        else:
            print u"创建模板失败:"
            print data
        return redirect(url_for('admin.application'))


@admin.route('/application/delete', methods=['POST'])
@admin_login
def application_delete():
    """管理员删除模板，前端传过来的模板id为数据库所存的模板主键而非k8s端的id"""
    cur_app = Application.query.filter_by(id=request.form['aid']).first_or_404()
    api_client = get_secure_api_client(session, current_user)
    if api_client.delete_app(cur_app.aid) == 200:
        db.session.delete(cur_app)
        db.session.commit()
        print u"删除模板成功"
        return jsonify(status="success")
    else:
        print u"删除模板失败"
        return jsonify(status="failed")


@admin.route('/application/<int:aid>/pic', methods=['GET', 'POST'])
@admin_login
def application_pic(aid):
    """上传模板图片"""
    cur_app = Application.query.filter_by(id=aid).first_or_404()
    if request.method == 'GET':
        return render_template('admin/application/picture.html',
                               title=u"模板封面管理",
                               app=cur_app)
    elif request.method == 'POST':
        _file = request.files['file']
        app_folder = current_app.config['APPLICATION_FOLDER']
        file_type = get_file_type(_file.mimetype)
        if _file and '.' in _file.filename and file_type == "img":
            im = Image.open(_file)
            # im.thumbnail((383, 262), Image.ANTIALIAS)
            im.resize((383, 262), Image.ANTIALIAS)
            image_path = os.path.join(app_folder, "%d.png" % cur_app.id)
            im.save(image_path, 'PNG')
            unique_mark = os.stat(image_path).st_mtime
            cur_app.cover_img = '/static/upload/application/' + '%d.png?t=%s' % (cur_app.id, unique_mark)
            db.session.commit()
            return redirect(url_for('admin.application'))
        else:
            message_fail = u"无效的文件类型"
            return render_template('admin/application/picture.html',
                                   title=u"模板封面管理",
                                   app=cur_app,
                                   message_fail=message_fail)


"""
    资讯与公告部分
"""


@admin.route('/news')
@admin_login
def news():
    """管理员查看资讯列表"""
    news_list = News.query.all()
    return render_template('admin/article/news.html',
                           title=u'新闻资讯管理',
                           news_list=news_list)


@admin.route('/news/create', methods=['GET', 'POST'])
@admin_login
def news_create():
    """管理员创建资讯"""
    if request.method == 'GET':
        return render_template('admin/article/news_create.html', title=u'创建新闻资讯')
    elif request.method == 'POST':
        _form = request.form
        title = _form['title']
        poster = _form['poster']
        content = _form['content'].replace("\n", "")
        new_news = News(title=title, poster=poster, content=content)
        db.session.add(new_news)
        db.session.commit()
        return redirect(url_for('admin.news'))


@admin.route('/news/edit/<int:nid>', methods=['GET', 'POST'])
@admin_login
def news_edit(nid):
    """管理员编辑资讯"""
    if request.method == 'GET':
        cur_news = News.query.filter_by(id=nid).first_or_404()
        return render_template('admin/article/news_edit.html', title=u'编辑新闻资讯', news=cur_news)
    elif request.method == 'POST':
        _form = request.form
        cur_news = News.query.filter_by(id=nid).first_or_404()
        cur_news.title = _form['title']
        cur_news.poster = _form['poster']
        cur_news.content = _form['content']
        db.session.commit()
        return redirect(url_for('admin.news'))


@admin.route('/news/delete', methods=['POST'])
@admin_login
def news_delete():
    """管理员删除资讯"""
    nid = request.form['nid']
    cur_news = News.query.filter_by(id=nid).first_or_404()
    db.session.delete(cur_news)
    db.session.commit()
    return jsonify(status="success")


@admin.route('/notice')
@admin_login
def notice():
    """管理员查看公告列表"""
    notice_list = Notice.query.all()
    return render_template('admin/article/notice.html',
                           title=u'系统公告管理',
                           notice_list=notice_list)


@admin.route('/notice/create', methods=['GET', 'POST'])
@admin_login
def notice_create():
    """管理员创建公告"""
    if request.method == 'GET':
        return render_template('admin/article/notice_create.html', title=u'创建系统公告')
    elif request.method == 'POST':
        _form = request.form
        title = _form['title']
        poster = _form['poster']
        content = _form['content'].replace("\n", "")
        new_notice = Notice(title=title, poster=poster, content=content)
        db.session.add(new_notice)
        db.session.commit()
        return redirect(url_for('admin.notice'))


@admin.route('/notice/edit/<int:nid>', methods=['GET', 'POST'])
@admin_login
def notice_edit(nid):
    """管理员编辑公告"""
    if request.method == 'GET':
        cur_notice = Notice.query.filter_by(id=nid).first_or_404()
        return render_template('admin/article/notice_edit.html', title=u'编辑系统公告', notice=cur_notice)
    elif request.method == 'POST':
        _form = request.form
        cur_notice = Notice.query.filter_by(id=nid).first_or_404()
        cur_notice.title = _form['title']
        cur_notice.poster = _form['poster']
        cur_notice.content = _form['content']
        db.session.commit()
        return redirect(url_for('admin.notice'))


@admin.route('/notice/delete', methods=['POST'])
@admin_login
def notice_delete():
    """管理员删除公告"""
    nid = request.form['nid']
    cur_notice = Notice.query.filter_by(id=nid).first_or_404()
    db.session.delete(cur_notice)
    db.session.commit()
    return jsonify(status="success")

