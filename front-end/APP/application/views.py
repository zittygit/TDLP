#! /usr/bin/env python
# -*- coding: utf-8 -*-
from flask import render_template, redirect, request, url_for, current_app, abort, jsonify, session
from flask_login import login_user, logout_user, current_user, login_required
from datetime import datetime
from ..util.API_manage import ApiClient
from ..util.authorize import user_auth
from ..models import Application
from . import application
from .. import db
import json


@application.route('/', methods=['GET', 'POST'])
@login_required
@user_auth
def index():
    apps = Application.query.all()
    for app in apps:
        app.param_list = json.loads(app.param)
    return render_template("application/index.html",
                           apps=apps,
                           title=u"应用模板")


