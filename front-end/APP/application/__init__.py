#! /usr/bin/env python
# -*- coding: utf-8 -*-


from flask import Blueprint
application = Blueprint('application', __name__)
from . import views
