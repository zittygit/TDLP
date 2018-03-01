#! /usr/bin/env python
# -*- coding: utf-8 -*-

"""
    本文件是项目本身的构造文件
    主要包括创建 flask app 的工厂函数
    配置 Flask 扩展插件时往往在工厂函数中对 app 进行相关的初始化。
"""

from flask import Flask
from flask_sqlalchemy import SQLAlchemy
from flask_login import LoginManager
from config import config

db = SQLAlchemy()
login_manager = LoginManager()
login_manager.session_protection = 'strong'  # 设置session保护级别
login_manager.login_view = 'user.login'     # 设置登录视图


def create_app(config_name):
    app = Flask(__name__)
    app.config.from_object(config[config_name])
    config[config_name].init_app(app)

    db.init_app(app)
    login_manager.init_app(app)

    # 注册路由
    from .util import filter_blueprint
    app.register_blueprint(filter_blueprint)
    from .main import main as main_blueprint
    app.register_blueprint(main_blueprint)
    from .admin import admin as admin_blueprint
    app.register_blueprint(admin_blueprint, url_prefix='/admin')
    from .user import user as user_blueprint
    app.register_blueprint(user_blueprint, url_prefix='/user')
    from .instance import instance as instance_blueprint
    app.register_blueprint(instance_blueprint, url_prefix='/instance')
    from .application import application as application_blueprint
    app.register_blueprint(application_blueprint, url_prefix='/application')
    from .article import article as article_blueprint
    app.register_blueprint(article_blueprint, url_prefix='/article')
    from .machine import machine as machine_blueprint
    app.register_blueprint(machine_blueprint, url_prefix='/machine')
    return app

