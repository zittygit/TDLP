#! /usr/bin/env python
# -*- coding: utf-8 -*-


from flask import Blueprint
instance = Blueprint('instance', __name__)
from . import views
